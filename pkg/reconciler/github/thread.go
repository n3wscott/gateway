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
	"fmt"
	"github.com/n3wscott/gateway/pkg/apis/gateway/v1alpha1"
	"knative.dev/pkg/apis"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/go-github/github"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	ghbinding "github.com/n3wscott/gateway/pkg/bindings/github"
	githubinformer "github.com/n3wscott/gateway/pkg/client/injection/informers/gateway/v1alpha1/github"
	listers "github.com/n3wscott/gateway/pkg/client/listers/gateway/v1alpha1"
)

type Instance interface {
	// Start starts the instance. Only call once.
	Start(ctx context.Context) error
}

func NewInstance(ctx context.Context, resync func()) Instance {
	return &githubInstance{
		resync: resync,
	}
}

type githubInstance struct {
	mtx sync.Mutex

	githubLister listers.GitHubLister
	dirty        bool

	resync func()

	client *github.Client

	ce      cloudevents.Client
	targets []string
}

var _ Instance = (*githubInstance)(nil)

// Start is a blocking call.
func (s *githubInstance) Start(ctx context.Context) error {
	githubInformer := githubinformer.Get(ctx)
	s.githubLister = githubInformer.Lister()

	githubInformer.Informer().AddEventHandler(controller.HandleAll(func(i interface{}) {
		s.mtx.Lock()
		s.dirty = true
		s.mtx.Unlock()
	}))

	c, err := ghbinding.New(ctx)
	if err != nil {
		return err
	}
	s.client = c

	if err := s.syncOrganization(ctx); err != nil {
		return err
	}

	// force a resync now.
	s.syncSinks(ctx)
	s.resync()

	outboundChan := make(chan cloudevents.Event, 20)

	if err := s.makeCloudEventsClient(ctx); err != nil {
		return err
	}

	heartbeat := time.NewTicker(time.Second * 600) // 5 minute heartbeat
	defer heartbeat.Stop()

	dirtyBit := time.NewTicker(time.Second * 5) // 5 second dirty bit check
	defer dirtyBit.Stop()
	for {
		select {
		case event := <-outboundChan:
			// This does a form of fanout. Tries once each target.
			for _, target := range s.targets {
				cectx := cloudevents.ContextWithTarget(ctx, target)
				if _, _, err := s.ce.Send(cectx, event); err != nil {
					logging.FromContext(cectx).
						With(zap.Error(err)).
						With(zap.String("target", target)).
						Warn("Failed to send to target.")
				}
			}

		case <-dirtyBit.C:
			s.mtx.Lock()
			wasDirty := s.dirty
			s.dirty = false
			s.mtx.Unlock()
			if wasDirty {
				s.syncSinks(ctx)
			}

		case <-heartbeat.C:
			logging.FromContext(ctx).Info("<3 heartbeat from github instance.")

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *githubInstance) makeCloudEventsClient(ctx context.Context) error {
	// TODO: this will not work because we want to share port 8080 at the top with slack and github. But for now we can
	// ignore the path.
	t, err := cloudevents.NewHTTPTransport(
		http.WithBinaryEncoding(),
		//http.WithPort(8080),
	)
	if err != nil {
		return err
	}

	if s.ce, err = cloudevents.NewClient(t,
		client.WithTimeNow(),
		client.WithUUIDs(),
	); err != nil {
		return err
	}

	//go func() {
	//	if err := s.ce.StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event) {
	//		// TODO: need to filter on identity as well?
	//		switch event.Type() {
	//		case events.ResponseEvent:
	//			resp := events.Message{}
	//			if err := event.DataAs(&resp); err != nil {
	//				logging.FromContext(ctx).With(zap.Error(err)).Error("Failed to get data from event.")
	//			}
	//			respChan <- resp
	//
	//		default:
	//			// skip sending to the channel.
	//		}
	//	}); err != nil {
	//		logging.FromContext(ctx).With(zap.Error(err)).Error("Failed to start cloudevents receiver.")
	//	}
	//}()

	return nil
}

func (s *githubInstance) syncSinks(ctx context.Context) {
	ss, err := s.githubLister.List(labels.Everything())
	if err != nil {
		logging.FromContext(ctx).With(zap.Error(err)).Error("Could not sync the sinks.")
	}

	targets := sets.NewString()
	for _, s := range ss {
		if s.Status.SinkURI != nil {
			targets.Insert(s.Status.SinkURI.String())
		}
	}

	s.mtx.Lock()
	s.targets = targets.List()
	s.mtx.Unlock()
}

func (s *githubInstance) GetOrganization(ctx context.Context, gh *v1alpha1.GitHub) (*v1alpha1.GitHubOrganization, error) {
	// test to see if org is an Organization or a User

	org, resp, err := s.client.Organizations.Get(ctx, gh.Spec.Organization)
	if err != nil {
		// not a org?
	}

}
