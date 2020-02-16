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
	"context"
	listers "github.com/n3wscott/gateway/pkg/client/listers/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"
	"knative.dev/pkg/system"

	"github.com/n3wscott/gateway/pkg/apis/gateway/v1alpha1"
	reconcilergithub "github.com/n3wscott/gateway/pkg/client/injection/reconciler/gateway/v1alpha1/github"
)

const (
	serviceName = "gateway"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason SlackReconciled.
func newReconciledNormal(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "GitHubReconciled", "GitHub reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler reconciles a GitHub object
type Reconciler struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	kubeClientSet kubernetes.Interface

	// listers index properties about resources
	githubLister listers.GitHubLister

	sinkResolver *resolver.URIResolver

	gh Instance
}

// Check that our Reconciler implements Interface
var _ reconcilergithub.Interface = (*Reconciler)(nil)

func getAddressableDestination() duckv1.Destination {
	path, _ := apis.ParseURL("/github")
	return duckv1.Destination{
		Ref: &duckv1.KReference{
			Kind:       "Service",
			Namespace:  system.Namespace(),
			Name:       serviceName,
			APIVersion: "v1",
		},
		URI: path,
	}
}

func getSinkDestination(source *v1alpha1.GitHub) duckv1.Destination {
	dest := source.Spec.Sink.DeepCopy()
	if dest.Ref != nil {
		// To call URIFromDestination(), dest.Ref must have a Namespace. If there is
		// no Namespace defined in dest.Ref, we will use the Namespace of the source
		// as the Namespace of dest.Ref.
		if dest.Ref.Namespace == "" {
			//TODO how does this work with deprecated fields
			dest.Ref.Namespace = source.GetNamespace()
		}
	}
	return *dest
}

// TODO: leverage the instance name and instance annotations.

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, source *v1alpha1.GitHub) pkgreconciler.Event {
	source.Status.InitializeConditions()
	source.Status.ObservedGeneration = source.Generation

	// Mark Address
	addrURI, err := r.sinkResolver.URIFromDestinationV1(getAddressableDestination(), source)
	if err != nil {
		source.Status.MarkNoAddress("NotFound", "")
		return err
	}
	source.Status.MarkAddress(addrURI)

	// Mark Sink
	sinkURI, err := r.sinkResolver.URIFromDestinationV1(getSinkDestination(source), source)
	if err != nil {
		source.Status.MarkNoSink("NotFound", "")
		return err
	}
	source.Status.MarkSink(sinkURI)

	return newReconciledNormal(source.Namespace, source.Name)
}
