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
	"github.com/google/go-cmp/cmp"
	"github.com/n3wscott/gateway/pkg/apis/gateway/v1alpha1"
	"knative.dev/pkg/apis"
	"strings"
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
	GetOrganization(ctx context.Context, gh *v1alpha1.GitHub) (*v1alpha1.GitHubOrganization, error)
	GetRepositories(ctx context.Context, gh *v1alpha1.GitHub) (v1alpha1.GitHubRepositories, error)
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

	ce       cloudevents.Client
	targets  []string
	orgRepos []string

	knownPR map[string]*GitHubPRState
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

	// force a resync now.
	s.syncSinks(ctx)
	s.syncOrgRepos(ctx)
	s.resync()

	outboundChan := make(chan cloudevents.Event, 20)
	s.knownPR = make(map[string]*GitHubPRState, 0)

	if err := s.makeCloudEventsClient(ctx); err != nil {
		return err
	}

	poll := time.NewTicker(time.Second * 60) // 1 minute poll
	defer poll.Stop()

	heartbeat := time.NewTicker(time.Second * 600) // 5 minute heartbeat
	defer heartbeat.Stop()

	dirtyBit := time.NewTicker(time.Second * 5) // 5 second dirty bit check
	defer dirtyBit.Stop()
	for {
		select {
		case <-poll.C:
			s.poll(ctx, s.orgRepos, outboundChan)

		case event := <-outboundChan:
			// This does a form of fanout. Tries once each target.
			for _, target := range s.targets { // TODO: UPDATE THIS TOO.
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
				s.syncOrgRepos(ctx)
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
		if s.Status.IsReady() && s.Status.SinkURI != nil {
			targets.Insert(s.Status.SinkURI.String())
		}
	}

	s.mtx.Lock()
	s.targets = targets.List()
	s.mtx.Unlock()
}

func (s *githubInstance) syncOrgRepos(ctx context.Context) {
	ss, err := s.githubLister.List(labels.Everything())
	if err != nil {
		logging.FromContext(ctx).With(zap.Error(err)).Error("Could not sync the Org/Repos.")
	}

	orgRepos := sets.NewString()
	for _, s := range ss {
		if s.Status.IsReady() && s.Spec.Organization != "" && len(s.Spec.Repositories) > 0 {
			org := s.Spec.Organization
			for _, repo := range s.Spec.Repositories {
				orgRepos.Insert(fmt.Sprintf("%s/%s", strings.TrimSpace(org), strings.TrimSpace(repo)))
			}
		}
	}

	s.mtx.Lock()
	s.orgRepos = orgRepos.List()
	logging.FromContext(ctx).Infof("Gonna send for %s", s.orgRepos)
	s.mtx.Unlock()
}

func (s *githubInstance) GetOrganization(ctx context.Context, gh *v1alpha1.GitHub) (*v1alpha1.GitHubOrganization, error) {
	// test to see if org is an Organization or a User

	org, _, err := s.client.Organizations.Get(ctx, gh.Spec.Organization)
	if err != nil {
		var usr *github.User
		usr, _, err = s.client.Users.Get(ctx, gh.Spec.Organization)
		if usr != nil {
			avatar, _ := apis.ParseURL(usr.GetAvatarURL())
			site, _ := apis.ParseURL(usr.GetHTMLURL())
			return &v1alpha1.GitHubOrganization{
				Name:   usr.GetName(),
				ID:     usr.GetID(),
				Login:  usr.GetLogin(),
				Avatar: avatar,
				URL:    site,
				Email:  usr.GetEmail(),
				Type:   "User",
			}, nil
		}
	} else if org != nil {
		avatar, _ := apis.ParseURL(org.GetAvatarURL())
		site, _ := apis.ParseURL(org.GetHTMLURL())
		return &v1alpha1.GitHubOrganization{
			Name:   org.GetName(),
			ID:     org.GetID(),
			Login:  org.GetLogin(),
			Avatar: avatar,
			URL:    site,
			Email:  org.GetEmail(),
			Type:   "Organization",
		}, nil
	}

	return nil, err
}

func (s *githubInstance) GetRepositories(ctx context.Context, gh *v1alpha1.GitHub) (v1alpha1.GitHubRepositories, error) {
	rs, _, err := s.client.Repositories.List(ctx, "knative", nil)
	if err != nil {
		return nil, err
	}
	repos := make(v1alpha1.GitHubRepositories, 0, len(rs))
	for _, repo := range rs {
		if repo.Archived != nil && *repo.Archived {
			continue
		}

		u, _ := apis.ParseURL(repo.GetHTMLURL())
		gu, _ := apis.ParseURL(repo.GetGitURL())
		repos = append(repos, v1alpha1.GitHubRepository{
			ID:          repo.GetID(),
			Name:        repo.GetName(),
			Description: repo.GetDescription(),
			Branch:      repo.GetDefaultBranch(),
			URL:         u,
			GitURL:      gu,
		})
	}
	return repos, nil
}

func (s *githubInstance) poll(ctx context.Context, orgRepos []string, ce chan cloudevents.Event) {
	for _, orgRepo := range orgRepos {

		or := strings.Split(orgRepo, "/")
		if len(or) != 2 {
			continue
			// TODO: log error
		}
		org := or[0]
		repo := or[1]

		// TODO: list will not give PRs that have been closed by default. Might have to track down
		// PRs that got closed or merged if not in list and we think they should be.
		prs, resp, err := s.client.PullRequests.List(ctx, org, repo, nil)
		if resp != nil {
			fmt.Println("Updated quota:", resp.Rate.String())
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		for i, pr := range prs {
			state := &GitHubPRState{
				URL:       pr.GetURL(),
				CreatedAt: pr.GetCreatedAt(),
				UpdatedAt: pr.GetUpdatedAt(),
				ClosedAt:  pr.GetClosedAt(),
				MergedAt:  pr.GetMergedAt(),
			}

			if knownState, found := s.knownPR[pr.GetURL()]; found {
				if diff := cmp.Diff(state, knownState); diff != "" {
					fmt.Println("Change in PR!", pr.GetHTMLURL())
					fmt.Println(diff)
				}
			} else {
				fmt.Printf("%v. %v\n\t%s\n\t%s\n", i+1, pr.GetTitle(), pr.GetState(), pr.GetHTMLURL())
				fmt.Printf("\t%s\t%s\t%s\t%s\n", pr.GetCreatedAt(), pr.GetUpdatedAt(), pr.GetMergedAt(), pr.GetClosedAt())
			}
			if !state.ClosedAt.IsZero() || !state.MergedAt.IsZero() {
				delete(s.knownPR, pr.GetURL())
			} else {
				s.knownPR[pr.GetURL()] = state
			}
		}
	}
}

type GitHubPRState struct {
	URL       string    `json:"url,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	ClosedAt  time.Time `json:"closedAt,omitempty"`
	MergedAt  time.Time `json:"mergedAt,omitempty"`
}
