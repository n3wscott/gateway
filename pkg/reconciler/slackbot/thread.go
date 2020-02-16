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
	"encoding/json"
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	slackbotinformer "github.com/n3wscott/gateway/pkg/client/injection/informers/gateway/v1alpha1/slackbot"
	listers "github.com/n3wscott/gateway/pkg/client/listers/gateway/v1alpha1"
	"github.com/n3wscott/gateway/pkg/rawslack"
	"github.com/n3wscott/gateway/pkg/slackbot/events"
	"github.com/nlopes/slack"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"sync"
	"time"

	"github.com/n3wscott/gateway/pkg/apis/gateway/v1alpha1"
	"github.com/n3wscott/gateway/pkg/bindings/slackbot"
)

type Instance interface {
	// Start starts the instance. Only call once.
	Start(ctx context.Context) error
	GetTeam(ctx context.Context) (*v1alpha1.SlackTeamInfo, error)
	GetChannels(ctx context.Context) (v1alpha1.SlackChannels, error)
	GetIMs(ctx context.Context) (v1alpha1.SlackIMs, error)
}

func NewInstance(ctx context.Context, resync func()) *slackbotInstance {
	return &slackbotInstance{
		resync: resync,
	}
}

type slackbotInstance struct {
	mtx sync.Mutex

	slackbotLister listers.SlackbotLister
	dirty          bool

	team     *v1alpha1.SlackTeamInfo
	channels v1alpha1.SlackChannels
	ims      v1alpha1.SlackIMs
	resync   func()

	client *slack.Client

	ce      cloudevents.Client
	targets []string
}

var _ Instance = (*slackbotInstance)(nil)

// Start is a blocking call.
func (s *slackbotInstance) Start(ctx context.Context) error {
	slackbotInformer := slackbotinformer.Get(ctx)
	s.slackbotLister = slackbotInformer.Lister()

	slackbotInformer.Informer().AddEventHandler(controller.HandleAll(func(i interface{}) {
		s.mtx.Lock()
		s.dirty = true
		s.mtx.Unlock()
	}))

	c, err := slackbot.New(ctx)
	if err != nil {
		return err
	}
	s.client = c

	if err := s.syncTeam(ctx); err != nil {
		return err
	}

	if err := s.syncChannels(ctx); err != nil {
		return err
	}

	if err := s.syncIMs(ctx); err != nil {
		return err
	}

	// force a resync now.
	s.syncSinks(ctx)
	s.resync()

	outboundChan := make(chan cloudevents.Event, 20)
	responseChan := make(chan events.Message, 20)

	if err := s.makeCloudEventsClient(ctx, responseChan); err != nil {
		return err
	}

	go s.startRTM(ctx, outboundChan, responseChan)

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
			logging.FromContext(ctx).Info("<3 heartbeat from slackbot instance.")

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *slackbotInstance) makeCloudEventsClient(ctx context.Context, respChan chan events.Message) error {
	// TODO: this will not work because we want to share port 8080 at the top with slack and github. But for now we can
	// ignore the path.
	t, err := cloudevents.NewHTTPTransport(
		http.WithBinaryEncoding(),
		http.WithPort(8080),
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

	go func() {
		if err := s.ce.StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event) {
			// TODO: need to filter on identity as well?
			switch event.Type() {
			case events.ResponseEvent:
				resp := events.Message{}
				if err := event.DataAs(&resp); err != nil {
					logging.FromContext(ctx).With(zap.Error(err)).Error("Failed to get data from event.")
				}
				respChan <- resp

			default:
				// skip sending to the channel.
			}
		}); err != nil {
			logging.FromContext(ctx).With(zap.Error(err)).Error("Failed to start cloudevents receiver.")
		}
	}()

	return nil
}

func (s *slackbotInstance) syncSinks(ctx context.Context) {
	sbs, err := s.slackbotLister.List(labels.Everything())
	if err != nil {
		logging.FromContext(ctx).With(zap.Error(err)).Error("Could not sync the sinks.")
	}

	targets := sets.NewString()
	for _, sb := range sbs {
		if sb.Status.SinkURI != nil {
			targets.Insert(sb.Status.SinkURI.String())
		}
	}

	s.mtx.Lock()
	s.targets = targets.List()
	s.mtx.Unlock()
}

func (s *slackbotInstance) syncTeam(ctx context.Context) error {
	ti, err := s.client.GetTeamInfo()
	if err != nil {
		return err
	}

	u, err := apis.ParseURL(fmt.Sprintf("https://%s.slack.com/", ti.Domain))
	if err != nil {
		return err
	}

	s.mtx.Lock()
	s.team = &v1alpha1.SlackTeamInfo{
		ID:   ti.ID,
		Name: ti.Name,
		URL:  u,
	}
	s.mtx.Unlock()

	return nil
}

func (s *slackbotInstance) GetTeam(ctx context.Context) (*v1alpha1.SlackTeamInfo, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.team == nil {
		return nil, errors.New("Team not set.")
	}
	return s.team.DeepCopy(), nil
}

func (s *slackbotInstance) syncChannels(ctx context.Context) error {
	chs, err := s.client.GetChannels(true)
	if err != nil {
		return err
	}

	channels := make(v1alpha1.SlackChannels, len(chs))
	for i, ch := range chs {
		channels[i] = v1alpha1.SlackChannel{
			Name:     ch.Name,
			ID:       ch.ID,
			IsMember: ch.IsMember,
		}
	}

	s.mtx.Lock()
	s.channels = channels
	s.mtx.Unlock()

	return nil
}

func (s *slackbotInstance) GetChannels(ctx context.Context) (v1alpha1.SlackChannels, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.channels == nil {
		return nil, errors.New("Channels not set.")
	}
	return s.channels.DeepCopy(), nil
}

func (s *slackbotInstance) syncIMs(ctx context.Context) error {
	imcs, err := s.client.GetIMChannels()
	if err != nil {
		return err
	}

	ims := make(v1alpha1.SlackIMs, 0, len(imcs))
	for _, imc := range imcs {
		if imc.IsUserDeleted {
			continue
		}

		ims = append(ims, v1alpha1.SlackIM{
			ID:   imc.ID,
			With: imc.User,
		})
	}

	s.mtx.Lock()
	s.ims = ims
	s.mtx.Unlock()

	return nil
}

func (s *slackbotInstance) GetIMs(ctx context.Context) (v1alpha1.SlackIMs, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.ims == nil {
		return nil, errors.New("IMs not set.")
	}
	return s.ims.DeepCopy(), nil
}

func (s *slackbotInstance) startRTM(ctx context.Context, ce chan cloudevents.Event, respChan chan events.Message) {
	logger := logging.FromContext(ctx).With(zap.String("method", "startRTM"))
	rtm := rawslack.NewRTM(s.client)
	go rtm.ManageConnection()

	go func() {
		for {
			select {
			case resp := <-respChan:
				logger.Infof("Sending %q to %s", resp.Text, resp.Channel)
				// TODO: I think new outgoing message makes an ID and that is how you can respond to a previous message.
				rtm.SendMessage(rtm.NewOutgoingMessage(resp.Text, resp.Channel))
			}
		}
	}()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *rawslack.ConnectingEvent:
			logger.Info("connecting")

		case *rawslack.ConnectedEvent:
			logger.Info("connected")

		case *rawslack.HelloEvent:
			logger.Info("hello")

		case *rawslack.LatencyReport:
			logger.Debug("latency")

		case *rawslack.AckMessage:
			logger.Infof("ack %+v", ev)

		case rawslack.RTMError:
			logger.Infof("error %+v", msg)

		case json.RawMessage:
			logger.Infof("%s: %v", msg.Type, string(ev))
			event := cloudevents.NewEvent(cloudevents.VersionV1)
			event.SetDataContentType(cloudevents.ApplicationJSON)
			event.SetType(events.Slack.Type(msg.Type))
			event.SetSource("slackbot.gateway.todo")
			_ = event.SetData(ev)
			ce <- event

		default:
			logger.Infof("Unexpected(%T): %v", msg.Data, msg.Data)
		}
	}
}
