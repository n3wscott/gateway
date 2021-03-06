/*
Copyright 2020 The Knative Authors.

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
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/webhook/resourcesemantics"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type GitHub struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the GitHub (from the client).
	Spec GitHubSpec `json:"spec"`

	// Status communicates the observed state of the GitHub (from the controller).
	// +optional
	Status GitHubStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *GitHub) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("GitHub")
}

var _ resourcesemantics.GenericCRD = (*GitHub)(nil)

// Check that GitHub is a runtime.Object.
var _ runtime.Object = (*GitHub)(nil)

// Check that we can create OwnerReferences to a GitHub.
var _ kmeta.OwnerRefable = (*GitHub)(nil)

// Check that GitHub implements the Conditions duck type.
var _ = duck.VerifyType(&GitHub{}, &duckv1.Conditions{})

// GitHubSpec holds the desired state of the GitHub (from the client).
type GitHubSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// +required
	Organization string `json:"org"`

	// +optional
	Repositories []string `json:"repos,omitempty"`
}

// GitHubStatus communicates the observed state of the GitHub (from the controller).
type GitHubStatus struct {
	// inherits duck/v1 SourceStatus, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last
	//   processed by the controller.
	// * Conditions - the latest available observations of a resource's current
	//   state.
	// * SinkURI - the current active sink URI that has been configured for the
	//   Source.
	duckv1.SourceStatus `json:",inline"`

	// AddressStatus is the part where the GitHub fulfills the Addressable contract.
	duckv1.AddressStatus `json:",inline"`

	Organization *GitHubOrganization `json:"org,omitempty"`

	Repositories GitHubRepositories `json:"repos,omitempty"`
}

type GitHubOrganization struct {
	Name   string    `json:"name,omitempty"`
	ID     int64     `json:"id,omitempty"`
	Login  string    `json:"login,omitempty"`
	Avatar *apis.URL `json:"avatar,omitempty"`
	URL    *apis.URL `json:"url,omitempty"`
	Email  string    `json:"email,omitempty"`
	Type   string    `json:"type,omitempty"` // Organization, User or Error
}

type GitHubRepository struct {
	ID          int64     `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Branch      string    `json:"branch,omitempty"` // default branch
	URL         *apis.URL `json:"url,omitempty"`
	GitURL      *apis.URL `json:"gitUrl,omitempty"`
}

type GitHubRepositories []GitHubRepository

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitHubList is a list of GitHub resources
type GitHubList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []GitHub `json:"items"`
}

// TODO: implement
func (in *GitHub) SetDefaults(context.Context) {
	// TODO: add default instance annotations.
}

// TODO: implement
func (in *GitHub) Validate(context.Context) *apis.FieldError {
	return nil
}
