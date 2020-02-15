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
}

func NewInstance(ctx context.Context, resync func()) *slackbotInstance {
	return &slackbotInstance{
		resync: resync,
	}
}

// Start is a blocking call.
func (s *slackbotInstance) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("<3 heartbeat from slackbot instance.")
		case <-ctx.Done():
			return nil
		}
	}
}

type slackbotInstance struct {
	mtx sync.Mutex

	team     *v1alpha1.SlackTeamInfo
	channels v1alpha1.SlackChannels
	resync   func()
}

var _ Instance = (*slackbotInstance)(nil)

func (s *slackbotInstance) GetTeam(ctx context.Context) (*v1alpha1.SlackTeamInfo, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.team == nil {
		return nil, errors.New("Team not set.")
	}
	return s.team.DeepCopy(), nil
}
func (s *slackbotInstance) GetChannels(ctx context.Context) (v1alpha1.SlackChannels, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.channels == nil {
		return nil, errors.New("Channels not set.")
	}
	return s.channels.DeepCopy(), nil
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
