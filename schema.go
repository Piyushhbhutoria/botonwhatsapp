package main

import "github.com/Rhymen/go-whatsapp"

type Payload struct {
	Message payloadData `json:"message"`
}

type payloadData struct {
	Content        string `json:"content"`
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
}

type SendText struct {
	Receiver string `json:"to"`
	Message  string `json:"text"`
}

type sendBulkText struct {
	List []SendText `json:"list"`
}

type waHandler struct {
	c *whatsapp.Conn
}

type resp struct {
	Results results `json:"results"`
}

type results struct {
	Messages []messages `json:"messages"`
}

type messages struct {
	Content string `json:"content"`
}
