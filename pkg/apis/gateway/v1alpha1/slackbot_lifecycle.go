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
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// SlackbotConditionReady has status True when the Slackbot is ready to send events.
	SlackbotConditionReady = apis.ConditionReady

	// SlackbotConditionSinkProvided has status True when the Slackbot has been configured with a sink target.
	SlackbotConditionSinkProvided apis.ConditionType = "SinkProvided"

	// SlackbotConditionAddressable has status True when there is a service for posting to.
	SlackbotConditionAddressable apis.ConditionType = "Addressable"
)

var slackCondSet = apis.NewLivingConditionSet(
	SlackbotConditionSinkProvided,
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

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *SlackbotStatus) MarkNoAddress(reason, messageFormat string, messageA ...interface{}) {
	slackCondSet.Manage(s).MarkFalse(SlackbotConditionAddressable, reason, messageFormat, messageA...)
}

// IsReady returns true if the resource is ready overall.
func (s *SlackbotStatus) IsReady() bool {
	return slackCondSet.Manage(s).IsHappy()
}
