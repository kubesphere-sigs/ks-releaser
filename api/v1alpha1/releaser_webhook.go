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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var releaserlog = logf.Log.WithName("releaser-resource")

func (r *Releaser) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-devops-kubesphere-io-v1alpha1-releaser,mutating=true,failurePolicy=fail,sideEffects=None,groups=devops.kubesphere.io,resources=releasers,verbs=create;update,versions=v1alpha1,name=mreleaser.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Releaser{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Releaser) Default() {
	releaserlog.Info("default", "name", r.Name)

	if r.Spec.Phase == "" {
		r.Spec.Phase = PhaseDraft
	}

	for i, _ := range r.Spec.Repositories {
		repo := &r.Spec.Repositories[i]
		if repo.Provider == "" {
			repo.Provider = ProviderGitHub
		}
		if repo.Action == "" {
			repo.Action = ActionTag
		}
		if repo.Branch == "" {
			repo.Branch = "master"
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-devops-kubesphere-io-v1alpha1-releaser,mutating=false,failurePolicy=fail,sideEffects=None,groups=devops.kubesphere.io,resources=releasers,verbs=create;update,versions=v1alpha1,name=vreleaser.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Releaser{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Releaser) ValidateCreate() error {
	releaserlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Releaser) ValidateUpdate(old runtime.Object) error {
	releaserlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Releaser) ValidateDelete() error {
	releaserlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
