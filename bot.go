package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"go.mau.fi/whatsmeow/types/events"
)

const (
	sapEndpoint = "https://api.cai.tools.sap/build/v1/dialog"
)

func askBot(evt *events.Message) (*resp, error) {
	payload := strings.NewReader(`{"message": {"content":"` + evt.Message.GetConversation() + `","type":"text"}, "conversation_id": "` + evt.Info.SourceString() + `"}`)
	fmt.Println("payload ->", *payload)

	req, err := http.NewRequest("POST", sapEndpoint, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", os.Getenv("SAP"))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var temp resp
		err = json.Unmarshal(body, &temp)
		if err != nil {
			zlog.Infof("Error decoding body: %v\n", err)
		}
		return &temp, nil
	}
	return nil, fmt.Errorf("Bad request: %v", res.StatusCode)
}
