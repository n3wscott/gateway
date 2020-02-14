package slackbot

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"
)

var _ psbinding.Bindable = (*SlackbotBinding)(nil)

func (sb *SlackbotBinding) GetNamespace() string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetNamespace(namespace string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetName() string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetName(name string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetGenerateName() string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetGenerateName(name string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetUID() types.UID {
	panic("implement me")
}

func (sb *SlackbotBinding) SetUID(uid types.UID) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetResourceVersion() string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetResourceVersion(version string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetGeneration() int64 {
	panic("implement me")
}

func (sb *SlackbotBinding) SetGeneration(generation int64) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetSelfLink() string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetSelfLink(selfLink string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetCreationTimestamp() v1.Time {
	panic("implement me")
}

func (sb *SlackbotBinding) SetCreationTimestamp(timestamp v1.Time) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetDeletionTimestamp() *v1.Time {
	panic("implement me")
}

func (sb *SlackbotBinding) SetDeletionTimestamp(timestamp *v1.Time) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetDeletionGracePeriodSeconds() *int64 {
	panic("implement me")
}

func (sb *SlackbotBinding) SetDeletionGracePeriodSeconds(*int64) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetLabels() map[string]string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetLabels(labels map[string]string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetAnnotations() map[string]string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetAnnotations(annotations map[string]string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetFinalizers() []string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetFinalizers(finalizers []string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetOwnerReferences() []v1.OwnerReference {
	panic("implement me")
}

func (sb *SlackbotBinding) SetOwnerReferences([]v1.OwnerReference) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetClusterName() string {
	panic("implement me")
}

func (sb *SlackbotBinding) SetClusterName(clusterName string) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetManagedFields() []v1.ManagedFieldsEntry {
	panic("implement me")
}

func (sb *SlackbotBinding) SetManagedFields(managedFields []v1.ManagedFieldsEntry) {
	panic("implement me")
}

func (sb *SlackbotBinding) GroupVersionKind() schema.GroupVersionKind {
	panic("implement me")
}

func (sb *SlackbotBinding) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	panic("implement me")
}

func (sb *SlackbotBinding) GetObjectKind() schema.ObjectKind {
	panic("implement me")
}

func (sb *SlackbotBinding) DeepCopyObject() runtime.Object {
	panic("implement me")
}

func (sb *SlackbotBinding) GetObjectMeta() v1.Object {
	panic("implement me")
}

func (sb *SlackbotBinding) GetGroupVersionKind() schema.GroupVersionKind {
	panic("implement me")
}

func (sb *SlackbotBinding) GetSubject() tracker.Reference {
	panic("implement me")
}

func (sb *SlackbotBinding) GetBindingStatus() duck.BindableStatus {
	panic("implement me")
}
