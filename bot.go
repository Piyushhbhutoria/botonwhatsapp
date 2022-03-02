package main

import (
	"encoding/json"
	"strings"

	"go.mau.fi/whatsmeow/types/events"
)

func reply(evt *events.Message) {
	if !evt.Info.IsFromMe || evt.Info.IsGroup {
		return
	}

	message := evt.Message.GetConversation()
	from := evt.Info.Sender.User

	payload := strings.NewReader(`{"message": {"content":"` + message + `","type":"text"}, "conversation_id": "` + from + `"}`)

	data, err := postRequest(payload)
	if err != nil {
		log.Errorf("Error in post request: %v\n", err)
	}
	var temp resp
	err = json.Unmarshal([]byte(data), &temp)
	if err != nil {
		log.Errorf("Error decoding body: %v\n", err)
	}

	if len(temp.Results.Messages) > 0 && temp.Results.Messages[0].Content != "I trigger the fallback skill because I don't understand or I don't know what I'm supposed to do..." {
		log.Infof("response is: %s", temp.Results.Messages[0].Content)
		to := from
		reply := temp.Results.Messages[0].Content
		log.Infof("%v --> %s\nBot --> %v", to, message, reply)
		log.Infof("-------------------------------")
		// args := []string{"send", to, reply}
		// text(args)
	}
}
