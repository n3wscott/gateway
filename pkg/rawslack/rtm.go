package rawslack

import (
	"encoding/json"
	"github.com/nlopes/slack"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	websocketDefaultTimeout = 10 * time.Second
	defaultPingInterval     = 30 * time.Second
)

const (
	rtmEventTypeAck                 = ""
	rtmEventTypeHello               = "hello"
	rtmEventTypeGoodbye             = "goodbye"
	rtmEventTypePong                = "pong"
	rtmEventTypeDesktopNotification = "desktop_notification"
)

// RTMOption options for the managed RTM.
type RTMOption func(*RTM)

// RTMOptionUseStart as of 11th July 2017 you should prefer setting this to false, see:
// https://api.slack.com/changelog/2017-04-start-using-rtm-connect-and-stop-using-rtm-start
func RTMOptionUseStart(b bool) RTMOption {
	return func(rtm *RTM) {
		rtm.useRTMStart = b
	}
}

// RTMOptionDialer takes a gorilla websocket Dialer and uses it as the
// Dialer when opening the websocket for the RTM connection.
func RTMOptionDialer(d *websocket.Dialer) RTMOption {
	return func(rtm *RTM) {
		rtm.dialer = d
	}
}

// RTMOptionPingInterval determines how often to deliver a ping message to slack.
func RTMOptionPingInterval(d time.Duration) RTMOption {
	return func(rtm *RTM) {
		rtm.pingInterval = d
		rtm.resetDeadman()
	}
}

// RTMOptionConnParams installs parameters to embed into the connection URL.
func RTMOptionConnParams(connParams url.Values) RTMOption {
	return func(rtm *RTM) {
		rtm.connParams = connParams
	}
}

// NewRTM returns a RTM, which provides a fully managed connection to
// Slack's websocket-based Real-Time Messaging protocol.
func NewRTM(api *slack.Client, options ...RTMOption) *RTM {
	result := &RTM{
		Client:           *api,
		IncomingEvents:   make(chan RTMEvent, 50),
		outgoingMessages: make(chan slack.OutgoingMessage, 20),
		pingInterval:     defaultPingInterval,
		pingDeadman:      time.NewTimer(deadmanDuration(defaultPingInterval)),
		killChannel:      make(chan bool),
		disconnected:     make(chan struct{}),
		disconnectedm:    &sync.Once{},
		forcePing:        make(chan bool),
		RawEvents:        make(chan json.RawMessage),
		idGen:            slack.NewSafeID(1),
		mu:               &sync.Mutex{},
	}

	for _, opt := range options {
		opt(result)
	}

	return result
}
