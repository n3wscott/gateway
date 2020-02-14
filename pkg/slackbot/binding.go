package slackbot

import (
	"context"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/webhook/psbinding"
	"path/filepath"

	"github.com/nlopes/slack"
)

const (
	VolumeName = "slack-bot-binding"
	MountPath  = "/var/bindings/slackbot"
)

// ReadKey may be used to read keys from the secret bound by the SlackBinding.
func ReadKey(key string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(MountPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AccessToken reads the file named accessToken that is mounted by the SlackBinding.
func AccessToken() (string, error) {
	return ReadKey("token")
}

// New instantiates a new github client from the access token from the SlackBinding
func New(ctx context.Context) (*slack.Client, error) {
	at, err := AccessToken()
	if err != nil {
		return nil, err
	}
	return slack.New(at), nil
}

func NewBinding(secret corev1.LocalObjectReference) psbinding.Bindable {
	return &SlackbotBinding{Secret: secret}
}

type SlackbotBinding struct {
	Secret corev1.LocalObjectReference
}

func (sb *SlackbotBinding) Do(ctx context.Context, ps *duckv1.WithPod) {

	// First undo so that we can just unconditionally append below.
	sb.Undo(ctx, ps)

	// Make sure the PodSpec has a Volume like this:
	volume := corev1.Volume{
		Name: VolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: sb.Secret.Name,
			},
		},
	}
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, volume)

	// Make sure that each [init]container in the PodSpec has a VolumeMount like this:
	volumeMount := corev1.VolumeMount{
		Name:      VolumeName,
		ReadOnly:  true,
		MountPath: MountPath,
	}
	spec := ps.Spec.Template.Spec
	for i := range spec.InitContainers {
		spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts, volumeMount)
	}
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMount)
	}
}

func (sb *SlackbotBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	// Make sure the PodSpec does NOT have the slack volume.
	for i, v := range spec.Volumes {
		if v.Name == VolumeName {
			ps.Spec.Template.Spec.Volumes = append(spec.Volumes[:i], spec.Volumes[i+1:]...)
			break
		}
	}

	// Make sure that none of the [init]containers have the slack volume mount
	for i, c := range spec.InitContainers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == VolumeName {
				spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts[:j], spec.InitContainers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}
	for i, c := range spec.Containers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == VolumeName {
				spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts[:j], spec.Containers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}
}
