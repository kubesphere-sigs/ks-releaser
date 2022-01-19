/*
Copyright 2021 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"path"
	"sigs.k8s.io/yaml"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
)

// ReleaserReconciler reconciles a Releaser object
type ReleaserReconciler struct {
	logger logr.Logger
	client.Client
	Scheme      *runtime.Scheme
	GitCacheDir string

	gitUser string
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;watch;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Releaser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ReleaserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	r.logger = log.FromContext(ctx)

	releaser := &devopsv1alpha1.Releaser{}
	if err = r.Get(ctx, req.NamespacedName, releaser); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}
	spec := releaser.Spec
	if spec.Phase != devopsv1alpha1.PhaseReady {
		return
	}

	if !r.needToUpdate(ctx, releaser) {
		return
	}

	r.logger.Info("start to release", "name", releaser.Name)
	secret := &v1.Secret{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: spec.Secret.Namespace,
		Name:      spec.Secret.Name,
	}, secret); err != nil {
		return
	}

	releaser.Status.Conditions = make([]devopsv1alpha1.Condition, 0)
	r.gitUser = string(secret.Data[v1.BasicAuthUsernameKey])

	var errSlice = ErrorSlice{}
	for i, _ := range spec.Repositories {
		repo := spec.Repositories[i]
		releaseRrr := release(repo, secret, r.gitUser)

		var condition devopsv1alpha1.Condition
		if releaseRrr == nil {
			condition = devopsv1alpha1.Condition{
				ConditionType: devopsv1alpha1.ConditionTypeRelease,
				Status:        devopsv1alpha1.ConditionStatusSuccess,
				Message:       fmt.Sprintf("%s was released", repo.Address),
			}
		} else {
			errSlice = errSlice.append(releaseRrr)
			condition = devopsv1alpha1.Condition{
				ConditionType: devopsv1alpha1.ConditionTypeRelease,
				Status:        devopsv1alpha1.ConditionStatusFailed,
				Message:       fmt.Sprintf("failed to release %s, error: %v", repo.Address, releaseRrr.Error()),
			}
		}
		addCondition(releaser, condition)
	}

	if err = errSlice.ToError(); err == nil {
		if err = r.markAsDone(secret, releaser); err != nil {
			condition := devopsv1alpha1.Condition{
				ConditionType: devopsv1alpha1.ConditionTypeOther,
				Status:        devopsv1alpha1.ConditionStatusFailed,
				Message:       fmt.Sprintf("failed to mark as done: %v", err),
			}
			addCondition(releaser, condition)
		}
	}

	if err == nil {
		releaser.Status.CompletionTime = &metav1.Time{Time: time.Now()}
		metricsRecord(releaser.DeepCopy())
	}

	if updateErr := r.Status().Update(ctx, releaser); err == nil && updateErr == nil {
		err = r.updateHash(ctx, releaser)
	} else if err == nil && updateErr != nil {
		err = updateErr
	}

	if err != nil {
		result = ctrl.Result{
			RequeueAfter: time.Second * 5,
		}
	}
	return
}

func (r *ReleaserReconciler) needToUpdate(ctx context.Context, releaser *devopsv1alpha1.Releaser) bool {
	hash := releaser.Annotations["releaser.devops.kubesphere.io/hash"]
	newHash := ComputeHash(releaser.Spec)
	if hash == "" {
		_ = r.updateHash(ctx, releaser)
		return true
	}
	return hash != newHash
}

func (r *ReleaserReconciler) updateHash(ctx context.Context, releaser *devopsv1alpha1.Releaser) (err error) {
	newHash := ComputeHash(releaser.Spec)
	if releaser.Annotations == nil {
		releaser.Annotations = make(map[string]string)
	}
	releaser.Annotations["releaser.devops.kubesphere.io/hash"] = newHash
	err = r.Update(ctx, releaser)
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReleaserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1alpha1.Releaser{}).
		Complete(r)
}

func (r *ReleaserReconciler) markAsDone(secret *v1.Secret, releaser *devopsv1alpha1.Releaser) (err error) {
	gitOps := releaser.Spec.GitOps
	if gitOps == nil || !gitOps.Enable {
		releaser.Spec.Phase = devopsv1alpha1.PhaseDone
		if err = r.Update(context.TODO(), releaser); err == nil {
			nextReleaser := releaser.DeepCopy()
			bumpReleaser(nextReleaser)

			if err = r.Create(context.TODO(), nextReleaser); err != nil {
				err = fmt.Errorf("failed to create next releaser: %s, error: %v", nextReleaser.GetName(), err)
			}
		}
		return
	}

	// use the secret from GitOps if it is existing
	if gitOps.Secret.Name != "" {
		namespacedName := types.NamespacedName{
			Namespace: gitOps.Secret.Namespace,
			Name:      gitOps.Secret.Name,
		}
		if err = r.Get(context.TODO(), namespacedName, secret); err != nil {
			err = fmt.Errorf("failed to find secret from %v, error: %v", namespacedName, err)
			return
		}
	}

	var gitRepo *git.Repository
	repo := gitOps.Repository
	if gitRepo, err = clone(repo.Address, repo.Branch, getAuth(secret), r.GitCacheDir); err != nil {
		err = fmt.Errorf("failed to clone repository: %s, error: %v", repo.Address, err)
		return
	}

	var gitRepoURL *url.URL
	if gitRepoURL, err = url.Parse(repo.Address); err != nil {
		return
	}

	repoDir := path.Join(r.GitCacheDir, gitRepoURL.Path)
	currentReleaserPath := findReleaserFile(fmt.Sprintf("%s.yaml", releaser.Name), repoDir)

	var data []byte
	copiedReleaser := releaser.DeepCopy()
	copiedReleaser.Spec.Phase = devopsv1alpha1.PhaseDone
	copiedReleaser.ObjectMeta.ManagedFields = nil
	copiedReleaser.ObjectMeta.UID = ""
	copiedReleaser.ObjectMeta.ResourceVersion = ""
	data, _ = yaml.Marshal(copiedReleaser)

	r.logger.Info("start to commit phase to be done", "name", releaser.Name)
	if err = saveAndPush(gitRepo, r.gitUser, currentReleaserPath, data, secret,
		fmt.Sprintf("release %s", releaser.Name)); err != nil {
		err = fmt.Errorf("failed to write file %s, error: %v", currentReleaserPath, err)
		return
	}

	r.logger.Info("start to create next release file")
	var bumpFilename string
	if data, bumpFilename, err = bumpReleaserAsData(data); err != nil {
		err = fmt.Errorf("failed to bump releaser: %s, error: %v", currentReleaserPath, err)
	} else {
		bumpFilePath := path.Join(path.Dir(currentReleaserPath), bumpFilename)
		err = saveAndPush(gitRepo, r.gitUser, bumpFilePath, data, secret,
			fmt.Sprintf("prepare the next release of %s", releaser.Name))
	}
	return
}

func metricsRecord(releaser *devopsv1alpha1.Releaser) {
	increaseTagActionCount()

	if releaser.Spec.GitOps != nil && releaser.Spec.GitOps.Enable {
		increaseGitOpsCount()
	}

	for i, _ := range releaser.Spec.Repositories {
		repo := releaser.Spec.Repositories[i]
		switch repo.Action {
		case devopsv1alpha1.ActionPreRelease:
			increasePreReleaseActionCount()
		case devopsv1alpha1.ActionRelease:
			increaseReleaseActionCount()
		}
	}
}

// addCondition adds or replaces a condition
func addCondition(releaser *devopsv1alpha1.Releaser, condition devopsv1alpha1.Condition) {
	releaser.Status.Conditions = append(releaser.Status.Conditions, condition)
}
