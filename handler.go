package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Rhymen/go-whatsapp"
)

func (*waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	// if message.Info.Timestamp > uint64(now) && !message.Info.FromMe && len(message.Text) < 17 && len(message.Text) > 1 {
	if message.Info.Timestamp > uint64(now) && !message.Info.FromMe && !strings.Contains(message.Text, "@g.us") && len(message.Text) < 21 {
		fmt.Printf("Received %s from %s\n", message.Text, message.Info.SenderJid)
		requestChannel <- message
	}
}

func (h *waHandler) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v/n", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		if err := h.c.Restore(); err != nil {
			log.Printf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}
