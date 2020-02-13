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

package slackbot

import (
	"context"
	"github.com/n3wscott/gateway/pkg/client/injection/reconciler/gateway/v1alpha1/slackbot"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/tools/cache"

	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"

	samplesourceClient "github.com/n3wscott/gateway/pkg/client/injection/client"
	slackbotinformer "github.com/n3wscott/gateway/pkg/client/injection/informers/gateway/v1alpha1/slackbot"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	deploymentInformer := deploymentinformer.Get(ctx)
	slackbotInformer := slackbotinformer.Get(ctx)

	r := &Reconciler{
		kubeClientSet:         kubeclient.Get(ctx),
		slackbotLister:    slackbotInformer.Lister(),
		deploymentLister:      deploymentInformer.Lister(),
		samplesourceClientSet: samplesourceClient.Get(ctx),
	}
	if err := envconfig.Process("", r); err != nil {
		logging.FromContext(ctx).Panicf("required environment variable is not defined: %v", err)
	}

	impl := slackbot.NewImpl(ctx, r)
	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	logging.FromContext(ctx).Info("Setting up event handlers")
	slackInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Slackbot")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
