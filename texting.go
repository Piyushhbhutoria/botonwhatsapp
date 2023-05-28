package main

import (
	"context"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func texting(to, mess string) string {
	recipient, ok := parseJID(to)
	if !ok {
		return "Error in JID"
	}
	check := checkuser(to)
	if !check {
		wLog.Errorf("User doesn't exist: %v", to)
		return "User doesn't exist"
	}

	msg := &waProto.Message{Conversation: proto.String(mess)}
	ts, err := cli.SendMessage(context.Background(), recipient, msg)
	var respstr string
	if err != nil {
		wLog.Errorf("Error sending message: %v", err)
		respstr = "Error sending message: " + err.Error()
	} else {
		wLog.Infof("Message sent (server timestamp: %s)", ts)
		respstr = "Message Sent -> " + to + " : " + ts.ID
	}
	return respstr
}
