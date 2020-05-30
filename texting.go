package main

import (
	"log"

	"github.com/Rhymen/go-whatsapp"
)

func texting(to, mess string) string {
	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: "91" + to + "@s.whatsapp.net",
		},
		Text: mess,
	}

	msgId, err := wac.Send(msg)
	if err != nil {
		log.Printf("Error sending message to %v --> %v\n", to, err)
		return "Error"
	}
	return "Message Sent -> " + to + " : " + msgId
}
