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

// Package adapter implements a sample receive adapter that generates events
// at a regular interval.
package slackbot

import (
	"context"
	"knative.dev/pkg/source"
	"log"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/n3wscott/gateway/pkg/slackbot"
	"knative.dev/eventing/pkg/adapter"
)

type envConfig struct {
	// Include the standard adapter.EnvConfig used by all adapters.
	adapter.EnvConfig
}

func NewEnv() adapter.EnvConfigAccessor { return &envConfig{} }

// Adapter generates events at a regular interval.
type Adapter struct {
	client cloudevents.Client
}

// Start runs the adapter.
// Returns if stopCh is closed or Send() returns an error.
func (a *Adapter) Start(stopCh <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())

	bot, err := slackbot.NewBot(ctx)
	if err != nil {
		log.Fatalf("failed to create bot: %s", err.Error())
	}

	go func() {
		if err := bot.Start(ctx); err != nil {
			log.Fatalf("failed to start cloudevent reciever: %s", err.Error())
		}
	}()

	<-stopCh
	cancel()
	return nil
}

func NewAdapter(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client, reporter source.StatsReporter) adapter.Adapter {
	env := aEnv.(*envConfig) // Will always be our own envConfig type
	_ = env
	return &Adapter{
		client: ceClient,
	}
}
