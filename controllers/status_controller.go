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
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"text/template"
)

// StatusController is responsible for reporting errors to git provider
type StatusController struct {
	logger logr.Logger
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devops.kubesphere.io,resources=releasers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;watch;list

// Reconcile responsible for reporting the error status of a releaser
func (c *StatusController) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	c.logger = log.FromContext(ctx)

	releaser := &devopsv1alpha1.Releaser{}
	if err = c.Get(ctx, req.NamespacedName, releaser); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// only support this controller when the gitops feature is enabled
	if releaser.Spec.GitOps == nil || !releaser.Spec.GitOps.Enable {
		return
	}

	// do not report the status when the phase was done
	if releaser.Spec.Phase == devopsv1alpha1.PhaseDone {
		return
	}

	// only take care of those have errors
	failedConditions := getFailedConditions(releaser.Status.Conditions)
	if len(failedConditions) == 0 {
		return
	}

	if err = c.createOrUpdateIssue(failedConditions, releaser.DeepCopy()); err != nil {
		c.logger.Error(err, "failed to create/update an issue for %v", req.NamespacedName)
		result = ctrl.Result{
			Requeue: true,
		}
	}
	return
}

func (c *StatusController) createOrUpdateIssue(conditions []devopsv1alpha1.Condition, releaser *devopsv1alpha1.Releaser) (
	err error) {
	secretRef := releaser.GetGitOpsSecret()

	secret := &v1.Secret{}
	if err = c.Get(context.TODO(), types.NamespacedName{
		Namespace: secretRef.Namespace,
		Name:      secretRef.Name,
	}, secret); err != nil {
		c.logger.Error(err, "cannot found secret: %v", secretRef)
		return
	}

	repo := releaser.Spec.GitOps.Repository
	gitProvider := getGitProviderClient(repo, secret)
	if gitProvider == nil {
		c.logger.Info(fmt.Sprintf("failed to get the git provider of %s", repo.Address))
		return
	}

	var issueBody string
	if issueBody, err = errorReportRender(releaser); err == nil {
		err = gitProvider.CreateIssue(releaser.Name, issueBody)
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *StatusController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1alpha1.Releaser{}).
		Complete(r)
}

func errorReportRender(releaser *devopsv1alpha1.Releaser) (message string, err error) {
	var tpl *template.Template
	tpl, err = template.New("report").Funcs(template.FuncMap{
		"trim": strings.TrimSpace,
	}).Parse(`Errors found with releaser: {{.Name}}
|Type|Status|Message|
|---|---|---|
{{- range .Status.Conditions}}
|{{.ConditionType}}|{{.Status}}|{{trim .Message}}|
{{- end}}
`)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer([]byte{})
	if err = tpl.Execute(buffer, releaser); err == nil {
		message = buffer.String()
	}
	return
}

func getFailedConditions(conditions []devopsv1alpha1.Condition) []devopsv1alpha1.Condition {
	if len(conditions) == 0 || conditions == nil {
		return nil
	}

	result := make([]devopsv1alpha1.Condition, 0)
	for i, _ := range conditions {
		if conditions[i].Status != devopsv1alpha1.ConditionStatusFailed {
			continue
		}
		result = append(result, conditions[i])
	}
	return result
}
