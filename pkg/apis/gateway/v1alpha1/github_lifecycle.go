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
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// GitHubConditionReady has status True when the GitHub is ready to send events.
	GitHubConditionReady = apis.ConditionReady

	// GitHubConditionSinkProvided has status True when the GitHub has been configured with a sink target.
	GitHubConditionSinkProvided apis.ConditionType = "SinkProvided"

	// GitHubConditionAddressable has status True when there is a service for posting to.
	GitHubConditionAddressable apis.ConditionType = "Addressable"
)

var githubCondSet = apis.NewLivingConditionSet(
	GitHubConditionSinkProvided,
	GitHubConditionAddressable,
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *GitHubStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return slackCondSet.Manage(s).GetCondition(t)
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *GitHubStatus) InitializeConditions() {
	slackCondSet.Manage(s).InitializeConditions()
}

// MarkSinkWarnDeprecated sets the condition that the source has a sink configured and warns ref is deprecated.
func (s *GitHubStatus) MarkSinkWarnRefDeprecated(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		c := apis.Condition{
			Type:     GitHubConditionSinkProvided,
			Status:   corev1.ConditionTrue,
			Severity: apis.ConditionSeverityError,
			Message:  "Using deprecated object ref fields when specifying spec.sink. These will be removed in a future release. Update to spec.sink.ref.",
		}
		slackCondSet.Manage(s).SetCondition(c)
	} else {
		slackCondSet.Manage(s).MarkUnknown(GitHubConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkSink sets the condition that the source has a sink configured.
func (s *GitHubStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		slackCondSet.Manage(s).MarkTrue(GitHubConditionSinkProvided)
	} else {
		slackCondSet.Manage(s).MarkUnknown(GitHubConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *GitHubStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	slackCondSet.Manage(s).MarkFalse(GitHubConditionSinkProvided, reason, messageFormat, messageA...)
}

func (ss *GitHubStatus) MarkAddress(url *apis.URL) {
	if ss.Address == nil {
		ss.Address = &duckv1.Addressable{}
	}
	if url != nil {
		ss.Address.URL = url
		slackCondSet.Manage(ss).MarkTrue(GitHubConditionAddressable)
	} else {
		ss.Address.URL = nil
		slackCondSet.Manage(ss).MarkFalse(GitHubConditionAddressable, "ServiceUnavailable", "Service was not created.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *GitHubStatus) MarkNoAddress(reason, messageFormat string, messageA ...interface{}) {
	slackCondSet.Manage(s).MarkFalse(GitHubConditionAddressable, reason, messageFormat, messageA...)
}

// IsReady returns true if the resource is ready overall.
func (s *GitHubStatus) IsReady() bool {
	return slackCondSet.Manage(s).IsHappy()
}
