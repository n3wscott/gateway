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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Slack struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the Slack (from the client).
	Spec SlackSpec `json:"spec"`

	// Status communicates the observed state of the Slack (from the controller).
	// +optional
	Status SlackStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *Slack) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Slack")
}

// Check that Slack is a runtime.Object.
var _ runtime.Object = (*Slack)(nil)

// Check that we can create OwnerReferences to a Slack.
var _ kmeta.OwnerRefable = (*Slack)(nil)

// Check that Slack implements the Conditions duck type.
var _ = duck.VerifyType(&Slack{}, &duckv1.Conditions{})

// SlackSpec holds the desired state of the Slack (from the client).
type SlackSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// Interval is the time interval between events.
	//
	// The string format is a sequence of decimal numbers, each with optional
	// fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time
	// units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Interval string `json:"interval"`
}

// SlackStatus communicates the observed state of the Slack (from the controller).
type SlackStatus struct {
	// inherits duck/v1 SourceStatus, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last
	//   processed by the controller.
	// * Conditions - the latest available observations of a resource's current
	//   state.
	// * SinkURI - the current active sink URI that has been configured for the
	//   Source.
	duckv1.SourceStatus `json:",inline"`

	// AddressStatus is the part where the Slack fulfills the Addressable contract.
	duckv1.AddressStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SlackList is a list of Slack resources
type SlackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Slack `json:"items"`
}
