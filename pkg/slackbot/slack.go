package slackbot

import (
	"context"
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/n3wscott/gateway/pkg/rawslack"
	"github.com/n3wscott/gateway/pkg/slackbot/events"
	"github.com/nlopes/slack"
	"path"
)

func (s *Bot) manageRTM(ctx context.Context) {

	for msg := range s.rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {

		case *rawslack.ConnectingEvent:
			fmt.Printf("connecting\n")
		case *rawslack.ConnectedEvent:
			fmt.Printf("connected\n")
		case *rawslack.HelloEvent:
			fmt.Printf("hello\n")
		case *rawslack.LatencyReport:
			fmt.Printf("latency\n")

		case *rawslack.AckMessage:
			fmt.Printf("ack %+v\n", ev)

		case rawslack.RTMError:
			fmt.Printf("error %+v\n", msg)

		case json.RawMessage:
			fmt.Printf("%s: %v\n", msg.Type, string(ev))

			event := cloudevents.NewEvent(cloudevents.VersionV1)
			event.SetDataContentType(cloudevents.ApplicationJSON)
			event.SetType(events.Slack.Type(msg.Type))
			event.SetSource("slackbot.gateway.todo")
			_ = event.SetData(ev)

			if _, _, err := s.ce.Send(ctx, event); err != nil {
				fmt.Printf("failed to send cloudevent: %v\n", err)
			}

		default:

			fmt.Printf("Unexpected(%T): %v\n", msg.Data, msg.Data)
		}
	}
}

func (s *Bot) manageRTM_old(ctx context.Context) {
	if team, err := s.client.GetTeamInfo(); err == nil {
		fmt.Printf("Slack Team: %+v", team)
		s.domain = team.Domain
	}

	for msg := range s.rtm.IncomingEvents {
		fmt.Println("Event Received: ", msg.Type)

		event := cloudevents.NewEvent(cloudevents.VersionV1)
		event.SetDataContentType(cloudevents.ApplicationJSON)
		event.SetType(events.Slack.Type(msg.Type))

		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			u := events.Slack.SourceForDomain(s.domain)
			event.SetSource(u.String())
			if err := event.SetData(ev); err != nil {
				fmt.Printf("failed to build cloudevent: %v\n", err)
				return
			}

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:

			u := events.Slack.SourceForChannel(s.domain, ev.Channel)

			switch ev.SubType {
			case "message_changed":
				event.SetSubject(ev.SubMessage.Timestamp)
				event.SetType(fmt.Sprintf("%s.%s", event.Type(), "changed"))
				u.Path = path.Join(u.Path, ev.SubMessage.Edited.User)
			case "message_deleted":
				event.SetSubject(ev.DeletedTimestamp)
				event.SetType(fmt.Sprintf("%s.%s", event.Type(), "deleted"))
			case "unpinned_item":
				event.SetType(fmt.Sprintf("%s.%s", event.Type(), "unpinned"))
			default:
				event.SetSubject(ev.Timestamp)
				u.Path = path.Join(u.Path, ev.User)
			}

			event.SetSource(u.String())
			if err := event.SetData(ev); err != nil {
				fmt.Printf("failed to build cloudevent: %v\n", err)
				return
			}

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			u := events.Slack.SourceForDomain(s.domain)
			event.SetSource(u.String())
			if err := event.SetData(ev); err != nil {
				fmt.Printf("failed to build cloudevent: %v\n", err)
				return
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
			s.Err <- fmt.Errorf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			s.Err <- fmt.Errorf("invalid credentials")

		default:
			// Ignore other events..
			fmt.Printf("Unexpected: %v\n", msg.Data)
		}

		if event.Data != nil {
			if _, _, err := s.ce.Send(ctx, event); err != nil {
				fmt.Printf("failed to send cloudevent: %v\n", err)
			}
		}
	}
}
