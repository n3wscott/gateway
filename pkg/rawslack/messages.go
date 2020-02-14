package rawslack

import "github.com/nlopes/slack"

// NewOutgoingMessage prepares an OutgoingMessage that the user can
// use to send a message. Use this function to properly set the
// messageID.
func (rtm *RTM) NewOutgoingMessage(text string, channelID string, options ...RTMsgOption) *slack.OutgoingMessage {
	id := rtm.idGen.Next()
	msg := slack.OutgoingMessage{
		ID:      id,
		Type:    "message",
		Channel: channelID,
		Text:    text,
	}
	for _, option := range options {
		option(&msg)
	}
	return &msg
}

// NewSubscribeUserPresence prepares an OutgoingMessage that the user can
// use to subscribe presence events for the specified users.
func (rtm *RTM) NewSubscribeUserPresence(ids []string) *slack.OutgoingMessage {
	return &slack.OutgoingMessage{
		Type: "presence_sub",
		IDs:  ids,
	}
}

// NewTypingMessage prepares an OutgoingMessage that the user can
// use to send as a typing indicator. Use this function to properly set the
// messageID.
func (rtm *RTM) NewTypingMessage(channelID string) *slack.OutgoingMessage {
	id := rtm.idGen.Next()
	return &slack.OutgoingMessage{
		ID:      id,
		Type:    "typing",
		Channel: channelID,
	}
}

// RTMsgOption allows configuration of various options available for sending an RTM message
type RTMsgOption func(*slack.OutgoingMessage)

// RTMsgOptionTS sets thead timestamp of an outgoing message in order to respond to a thread
func RTMsgOptionTS(threadTimestamp string) RTMsgOption {
	return func(msg *slack.OutgoingMessage) {
		msg.ThreadTimestamp = threadTimestamp
	}
}

// RTMsgOptionBroadcast sets broadcast reply to channel to "true"
func RTMsgOptionBroadcast() RTMsgOption {
	return func(msg *slack.OutgoingMessage) {
		msg.ThreadBroadcast = true
	}
}
