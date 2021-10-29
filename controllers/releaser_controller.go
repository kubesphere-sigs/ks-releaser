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
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"golang.org/x/crypto/ssh"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"os"
	"path"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
)

// ReleaserReconciler reconciles a Releaser object
type ReleaserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Releaser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ReleaserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	_ = log.FromContext(ctx)

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

	//if err = r.Get(ctx, req.NamespacedName, releaser); err != nil {
	//	err = client.IgnoreNotFound(err)
	//	return
	//}
	secret := &v1.Secret{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: spec.Secret.Namespace,
		Name:      spec.Secret.Name,
	}, secret); err != nil {
		return
	}

	var errSlice = ErrorSlice{}
	copiedReleaser := releaser.DeepCopy()
	for i, _ := range spec.Repositories {
		repo := spec.Repositories[i]
		releaseRrr := release(spec.Repositories[i], secret)
		var condition devopsv1alpha1.Condition
		if releaseRrr == nil {
			condition = devopsv1alpha1.Condition{
				RepositoryName: repo.Name,
				Status:         "success",
				Message:        "success",
			}
		} else {
			errSlice = errSlice.append(releaseRrr)
			condition = devopsv1alpha1.Condition{
				RepositoryName: repo.Name,
				Status:         "failed",
				Message:        releaseRrr.Error(),
			}
		}
		addCondition(copiedReleaser, condition)
	}

	if err = errSlice.ToError(); err == nil {
		copiedReleaser.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	} else {
		result = ctrl.Result{
			RequeueAfter: time.Second * 5,
		}
	}

	r.markAsDone(secret, copiedReleaser)
	if updateErr := r.Status().Update(ctx, copiedReleaser); updateErr == nil {
		r.updateHash(ctx, copiedReleaser)
	} else {
		fmt.Println(updateErr)
	}
	return
}

// addCondition adds or replaces a condition
func addCondition(releaser *devopsv1alpha1.Releaser, condition devopsv1alpha1.Condition) {
	if condition.RepositoryName == "" {
		return
	}

	for i, _ := range releaser.Status.Conditions {
		item := releaser.Status.Conditions[i]
		if item.RepositoryName == condition.RepositoryName {
			releaser.Status.Conditions[i] = condition
			return
		}
	}
	releaser.Status.Conditions = append(releaser.Status.Conditions, condition)
}

func (r *ReleaserReconciler) needToUpdate(ctx context.Context, releaser *devopsv1alpha1.Releaser) bool {
	hash := releaser.Annotations["releaser.devops.kubesphere.io/hash"]
	newHash := ComputeHash(releaser.Spec)
	if hash == "" {
		r.updateHash(ctx, releaser)
		return true
	}
	return hash != newHash
}

func (r *ReleaserReconciler) updateHash(ctx context.Context, releaser *devopsv1alpha1.Releaser) {
	newHash := ComputeHash(releaser.Spec)
	releaser.Annotations["releaser.devops.kubesphere.io/hash"] = newHash
	_ = r.Update(ctx, releaser)
}

func release(repo devopsv1alpha1.Repository, secret *v1.Secret) (err error) {
	auth := getAuth(secret)

	var gitRepo *git.Repository
	if gitRepo, err = clone(repo.Address, repo.Branch, auth, "."); err != nil {
		return
	}

	if repo.Message == "" {
		repo.Message = "released by ks-releaser"
	}
	if _, err = setTag(gitRepo, repo.Version, repo.Message); err != nil {
		return
	}

	err = pushTags(gitRepo, auth)
	return
}

func getAuth(secret *v1.Secret) (auth transport.AuthMethod) {
	switch secret.Type {
	case v1.SecretTypeBasicAuth:
		auth = &githttp.BasicAuth{
			Username: string(secret.Data[v1.BasicAuthUsernameKey]),
			Password: string(secret.Data[v1.BasicAuthPasswordKey]),
		}
	case v1.SecretTypeSSHAuth:
		signer, _ := ssh.ParsePrivateKey(secret.Data[v1.SSHAuthPrivateKey])
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}
	return
}

func clone(gitRepo, branch string, auth transport.AuthMethod, cacheDir string) (repo *git.Repository, err error) {
	var gitRepoURL *url.URL
	if gitRepoURL, err = url.Parse(gitRepo); err != nil {
		return
	}

	dir := path.Join(cacheDir, gitRepoURL.Path)
	if ok, _ := PathExists(dir); ok {
		if repo, err = git.PlainOpen(dir); err == nil {
			var wd *git.Worktree

			if wd, err = repo.Worktree(); err == nil {
				if err = wd.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName(branch),
					Create: false,
					Force:  true,
				}); err != nil {
					err = fmt.Errorf("unable to checkout git branch: %s", branch)
					return
				}

				if err = wd.Pull(&git.PullOptions{
					Progress:      os.Stdout,
					ReferenceName: plumbing.NewBranchReferenceName(branch),
					Force:         true, // in case of the force pushing
					Auth:          auth,
				}); err != nil && err != git.NoErrAlreadyUpToDate {
					err = fmt.Errorf("failed to pull git repository '%s', error: %v", repo, err)
				} else {
					err = nil
				}
			}
		} else {
			err = fmt.Errorf("failed to open git local repository, error: %v", err)
		}
	} else {
		repo, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL:           gitRepo,
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			Progress:      os.Stdout,
			Auth:          auth,
		})
	}
	return
}

func tagExists(tag string, r *git.Repository) bool {
	tagFoundErr := "tag was found"
	tags, err := r.TagObjects()
	if err != nil {
		fmt.Printf("get tags error: %s\n", err)
		return false
	}
	res := false
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			res = true
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
	if err != nil && err.Error() != tagFoundErr {
		fmt.Printf("iterate tags error: %s\n", err)
		return false
	}
	return res
}

func setTag(r *git.Repository, tag, message string) (bool, error) {
	if tagExists(tag, r) {
		fmt.Printf("tag %s already exists\n", tag)
		return false, nil
	}
	fmt.Printf("Set tag %s\n", tag)
	h, err := r.Head()
	if err != nil {
		fmt.Printf("get HEAD error: %s\n", err)
		return false, err
	}
	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  "ks-releaser",
			Email: "linuxsuren@gmail.com",
			When:  time.Time{},
		},
		Message: message,
	})

	if err != nil {
		fmt.Printf("create tag error: %s\n", err)
		return false, err
	}
	return true, nil
}

func pushTags(r *git.Repository, auth transport.AuthMethod) (err error) {
	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		//RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       auth,
	}
	if err = r.Push(po); err != nil {
		if err == git.NoErrAlreadyUpToDate {
			fmt.Print("origin remote was up to date, no push done\n")
			err = nil
			return
		}
		err = fmt.Errorf("push to remote origin error: %s\n", err)
	}
	return
}

// PathExists checks if the target path exist or not
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReleaserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1alpha1.Releaser{}).
		Complete(r)
}

func (r *ReleaserReconciler) markAsDone(secret *v1.Secret, releaser *devopsv1alpha1.Releaser) {
	gitOps := releaser.Spec.GitOps
	if gitOps == nil || !gitOps.Enable {
		releaser.Spec.Phase = devopsv1alpha1.PhaseDone
		return
	}

	repo := gitOps.Repository
	if gitRepo, err := clone(repo.Address, repo.Branch, getAuth(secret), "tmp"); err == nil {
		var gitRepoURL *url.URL
		if gitRepoURL, err = url.Parse(repo.Address); err != nil {
			return
		}

		dir := path.Join(".", gitRepoURL.Path)
		filePath := path.Join(dir, fmt.Sprintf("%s.yaml", releaser.Name))

		if data, err := ioutil.ReadFile(filePath); err == nil {
			data, err = updateReleaserAsYAML(data, func(releaser *devopsv1alpha1.Releaser) {
				releaser.Spec.Phase = devopsv1alpha1.PhaseDone
			})
			if err == nil {
				if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
					fmt.Println("failed to write file", filePath)
				} else {
					if err := addAndCommit(gitRepo); err == nil {
						err = pushTags(gitRepo, getAuth(secret))
					}
				}
			}
		}
		fmt.Println(err)
	}
}

func addAndCommit(repo *git.Repository) (err error){
	var w *git.Worktree
	if w, err = repo.Worktree(); err == nil {
		_, _ = w.Add(".")
		var commit plumbing.Hash
		commit, err = w.Commit("example go-git commit", &git.CommitOptions{
			All: true,
			Author: &object.Signature{
				Name:  "John Doe",
				Email: "john@doe.org",
				When:  time.Now(),
			},
		})

		if err == nil {
			_, err = repo.CommitObject(commit)
		}
	}
	return
}
