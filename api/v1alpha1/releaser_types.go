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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReleaserSpec defines the desired state of Releaser
type ReleaserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase is the stage of a release request
	Phase        Phase              `json:"phase,omitempty"`
	Version      string             `json:"version,omitempty"`
	Repositories []Repository       `json:"repositories,omitempty"`
	GitOps       *GitOps            `json:"gitOps,omitempty"`
	Secret       v1.SecretReference `json:"secret,omitempty"`
}

// Phase is the stage of release request
type Phase string

const (
	// PhaseDraft allows user to modify
	PhaseDraft Phase = "draft"
	// PhaseReady indicates this request is ready to release
	PhaseReady Phase = "ready"
	// PhaseDone indicates this request was done
	PhaseDone Phase = "done"
)

// IsValid checks if this is valid
func (p Phase) IsValid() bool {
	switch p {
	case PhaseDraft, PhaseReady, PhaseDone:
		return true
	default:
		return false
	}
}

// Repository represents a git repository
type Repository struct {
	// +optional
	Name string `json:"name"`
	// +optional
	Provider Provider `json:"provider,omitempty"`
	Address  string   `json:"address"`
	// +optional
	Branch string `json:"branch,omitempty"`
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
	// +optional
	Action Action `json:"action,omitempty"`
}

// GetDefaultProvider returns the default git provider
func GetDefaultProvider(r *Repository) Provider {
	if r.Provider != "" {
		return r.Provider
	}

	address := r.Address
	if strings.HasPrefix(address, "https://github.com/") {
		return ProviderGitHub
	} else if strings.HasPrefix(address, "https://gitlab.com/") {
		return ProviderGitlab
	} else if strings.HasPrefix(address, "https://bitbucket.org/") {
		return ProviderBitbucket
	} else if strings.HasPrefix(address, "https://gitee.com/") {
		return ProviderGitee
	} else if strings.HasPrefix(address, "https://gitea.com/") {
		return ProviderGitea
	} else if address != "" {
		return ProviderUnknown
	}
	return ""
}

// GitOps indicates to integrate with GitOps
type GitOps struct {
	Enable     bool               `json:"enable,omitempty"`
	Repository Repository         `json:"repository,omitempty"`
	Secret     v1.SecretReference `json:"secret,omitempty"`
}

// Provider represents a git provider, such as: GitHub, Gitlab
type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderGitlab    Provider = "gitlab"
	ProviderBitbucket Provider = "bitbucket"
	ProviderGitee     Provider = "gitee"
	ProviderGitea     Provider = "gitea"
	ProviderUnknown   Provider = "unknown"
)

// Action indicates the action once the request phase to be ready
type Action string

const (
	ActionAuto       Action = "auto"
	ActionTag        Action = "tag"
	ActionPreRelease Action = "pre-release"
	ActionRelease    Action = "release"
)

// ReleaserStatus defines the observed state of Releaser
type ReleaserStatus struct {
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	Conditions     []Condition  `json:"conditions,omitempty"`
}

// Condition indicates the status of each git repositories
type Condition struct {
	ConditionType ConditionType   `json:"conditionType"`
	Status        ConditionStatus `json:"status"`
	Message       string          `json:"message"`
}

// ConditionType is the type of condition
type ConditionType string

const (
	ConditionTypeRelease ConditionType = "release"
	ConditionTypeOther   ConditionType = "other"
)

// ConditionStatus is the status of a condition
type ConditionStatus string

const (
	ConditionStatusSuccess ConditionStatus = "success"
	ConditionStatusFailed  ConditionStatus = "failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.spec.phase`,description="The phase of a Releaser"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`,description="The version of a Releaser"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="The age of a Releaser"

// Releaser is the Schema for the releasers API
type Releaser struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ReleaserSpec `json:"spec,omitempty"`
	// +optional
	Status ReleaserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReleaserList contains a list of Releaser
type ReleaserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Releaser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Releaser{}, &ReleaserList{})
}
