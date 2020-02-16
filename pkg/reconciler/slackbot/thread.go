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
	"errors"
	"fmt"
	"github.com/nlopes/slack"
	"knative.dev/pkg/apis"
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

	team     *v1alpha1.SlackTeamInfo
	channels v1alpha1.SlackChannels
	ims      v1alpha1.SlackIMs
	resync   func()

	client *slack.Client
}

var _ Instance = (*slackbotInstance)(nil)

// Start is a blocking call.
func (s *slackbotInstance) Start(ctx context.Context) error {
	c, err := slackbot.New(ctx)
	if err != nil {
		return err
	}
	s.client = c

	if err := s.syncTeam(ctx); err != nil {
		return nil
	}

	if err := s.syncChannels(ctx); err != nil {
		return nil
	}

	if err := s.syncIMs(ctx); err != nil {
		return nil
	}

	// force a resync now.
	s.resync()

	heartbeat := time.NewTicker(time.Second * 30)
	defer heartbeat.Stop()
	for {
		select {
		case <-heartbeat.C:
			fmt.Println("<3 heartbeat from slackbot instance.")
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *slackbotInstance) syncTeam(ctx context.Context) error {
	ti, err := s.client.GetTeamInfo()
	if err != nil {
		return err
	}

	url, err := apis.ParseURL(fmt.Sprintf("https://%s.slack.com/", ti.Domain))
	if err != nil {
		return err
	}

	s.mtx.Lock()
	s.team = &v1alpha1.SlackTeamInfo{
		ID:   ti.ID,
		Name: ti.Name,
		URL:  url,
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

func do() {
	ctx := context.Background()

	c, err := slackbot.New(ctx)
	if err != nil {
		panic(err)
	}

	channels, err := c.GetChannels(false)
	if err != nil {
		panic(err)
	}
	fmt.Println("Channels:")
	for _, ch := range channels {
		fmt.Println(ch.Name, ch.ID)
	}

	ti, err := c.GetTeamInfo()
	if err != nil {
		panic(err)
	}
	fmt.Println("Team info:", ti)
}
