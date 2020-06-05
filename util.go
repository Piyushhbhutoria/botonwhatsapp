package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PiyushhBhutoria/botonwhatsapp/config"
)

func createPayload(reqPayload interface{}) (payload []byte, err error) {
	typePayload := reqPayload.(Payload)
	payload, err = json.Marshal(typePayload)
	if err != nil {
		return
	}
	return
}

func postRequest(endpoint string, payload []byte) (string, error) {
	config := config.GetConfig()
	fmt.Println("payload ->", string(payload))
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", config.GetString("SAP"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", "api.cai.tools.sap")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}
	return "", fmt.Errorf("Bad request: %v", res.StatusCode)
}
