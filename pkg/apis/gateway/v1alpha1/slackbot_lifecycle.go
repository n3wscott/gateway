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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// SlackbotConditionReady has status True when the Slackbot is ready to send events.
	SlackbotConditionReady = apis.ConditionReady

	// SlackbotConditionSinkProvided has status True when the Slackbot has been configured with a sink target.
	SlackbotConditionSinkProvided apis.ConditionType = "SinkProvided"

	// SlackbotConditionDeployed has status True when the Slackbot has had it's deployment created.
	SlackbotConditionDeployed apis.ConditionType = "Deployed"

	SlackbotConditionAddressable apis.ConditionType = "Addressable"
)

var slackCondSet = apis.NewLivingConditionSet(
	SlackbotConditionSinkProvided,
	SlackbotConditionDeployed,
	SlackbotConditionAddressable,
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *SlackbotStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return slackCondSet.Manage(s).GetCondition(t)
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *SlackbotStatus) InitializeConditions() {
	slackCondSet.Manage(s).InitializeConditions()
}

// MarkSinkWarnDeprecated sets the condition that the source has a sink configured and warns ref is deprecated.
func (s *SlackbotStatus) MarkSinkWarnRefDeprecated(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		c := apis.Condition{
			Type:     SlackbotConditionSinkProvided,
			Status:   corev1.ConditionTrue,
			Severity: apis.ConditionSeverityError,
			Message:  "Using deprecated object ref fields when specifying spec.sink. These will be removed in a future release. Update to spec.sink.ref.",
		}
		slackCondSet.Manage(s).SetCondition(c)
	} else {
		slackCondSet.Manage(s).MarkUnknown(SlackbotConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkSink sets the condition that the source has a sink configured.
func (s *SlackbotStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		slackCondSet.Manage(s).MarkTrue(SlackbotConditionSinkProvided)
	} else {
		slackCondSet.Manage(s).MarkUnknown(SlackbotConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *SlackbotStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	slackCondSet.Manage(s).MarkFalse(SlackbotConditionSinkProvided, reason, messageFormat, messageA...)
}

// PropagateDeploymentAvailability uses the availability of the provided Deployment to determine if
// SlackbotConditionDeployed should be marked as true or false.
func (s *SlackbotStatus) PropagateDeploymentAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		slackCondSet.Manage(s).MarkTrue(SlackbotConditionDeployed)
	} else {
		// I don't know how to propagate the status well, so just give the name of the Deployment
		// for now.
		slackCondSet.Manage(s).MarkFalse(SlackbotConditionDeployed, "DeploymentUnavailable", "The Deployment '%s' is unavailable.", d.Name)
	}
}

func (ss *SlackbotStatus) MarkAddress(url *apis.URL) {
	if ss.Address == nil {
		ss.Address = &duckv1.Addressable{}
	}
	if url != nil {
		ss.Address.URL = url
		slackCondSet.Manage(ss).MarkTrue(SlackbotConditionAddressable)
	} else {
		ss.Address.URL = nil
		slackCondSet.Manage(ss).MarkFalse(SlackbotConditionAddressable, "ServiceUnavailable", "Service was not created.")
	}
}

// IsReady returns true if the resource is ready overall.
func (s *SlackbotStatus) IsReady() bool {
	return slackCondSet.Manage(s).IsHappy()
}
