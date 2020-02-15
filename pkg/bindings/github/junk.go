/*
Copyright 2020 The Knative Authors

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
package github

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"
)

var _ psbinding.Bindable = (*GitHubBinding)(nil)

func (sb *GitHubBinding) GetNamespace() string {
	panic("implement me")
}

func (sb *GitHubBinding) SetNamespace(namespace string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetName() string {
	panic("implement me")
}

func (sb *GitHubBinding) SetName(name string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetGenerateName() string {
	panic("implement me")
}

func (sb *GitHubBinding) SetGenerateName(name string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetUID() types.UID {
	panic("implement me")
}

func (sb *GitHubBinding) SetUID(uid types.UID) {
	panic("implement me")
}

func (sb *GitHubBinding) GetResourceVersion() string {
	panic("implement me")
}

func (sb *GitHubBinding) SetResourceVersion(version string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetGeneration() int64 {
	panic("implement me")
}

func (sb *GitHubBinding) SetGeneration(generation int64) {
	panic("implement me")
}

func (sb *GitHubBinding) GetSelfLink() string {
	panic("implement me")
}

func (sb *GitHubBinding) SetSelfLink(selfLink string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetCreationTimestamp() v1.Time {
	panic("implement me")
}

func (sb *GitHubBinding) SetCreationTimestamp(timestamp v1.Time) {
	panic("implement me")
}

func (sb *GitHubBinding) GetDeletionTimestamp() *v1.Time {
	panic("implement me")
}

func (sb *GitHubBinding) SetDeletionTimestamp(timestamp *v1.Time) {
	panic("implement me")
}

func (sb *GitHubBinding) GetDeletionGracePeriodSeconds() *int64 {
	panic("implement me")
}

func (sb *GitHubBinding) SetDeletionGracePeriodSeconds(*int64) {
	panic("implement me")
}

func (sb *GitHubBinding) GetLabels() map[string]string {
	panic("implement me")
}

func (sb *GitHubBinding) SetLabels(labels map[string]string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetAnnotations() map[string]string {
	panic("implement me")
}

func (sb *GitHubBinding) SetAnnotations(annotations map[string]string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetFinalizers() []string {
	panic("implement me")
}

func (sb *GitHubBinding) SetFinalizers(finalizers []string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetOwnerReferences() []v1.OwnerReference {
	panic("implement me")
}

func (sb *GitHubBinding) SetOwnerReferences([]v1.OwnerReference) {
	panic("implement me")
}

func (sb *GitHubBinding) GetClusterName() string {
	panic("implement me")
}

func (sb *GitHubBinding) SetClusterName(clusterName string) {
	panic("implement me")
}

func (sb *GitHubBinding) GetManagedFields() []v1.ManagedFieldsEntry {
	panic("implement me")
}

func (sb *GitHubBinding) SetManagedFields(managedFields []v1.ManagedFieldsEntry) {
	panic("implement me")
}

func (sb *GitHubBinding) GroupVersionKind() schema.GroupVersionKind {
	panic("implement me")
}

func (sb *GitHubBinding) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	panic("implement me")
}

func (sb *GitHubBinding) GetObjectKind() schema.ObjectKind {
	panic("implement me")
}

func (sb *GitHubBinding) DeepCopyObject() runtime.Object {
	panic("implement me")
}

func (sb *GitHubBinding) GetObjectMeta() v1.Object {
	panic("implement me")
}

func (sb *GitHubBinding) GetGroupVersionKind() schema.GroupVersionKind {
	panic("implement me")
}

func (sb *GitHubBinding) GetSubject() tracker.Reference {
	panic("implement me")
}

func (sb *GitHubBinding) GetBindingStatus() duck.BindableStatus {
	panic("implement me")
}
