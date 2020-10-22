package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PiyushhBhutoria/botonwhatsapp/config"
)

func postRequest(payload *strings.Reader) (string, error) {
	config := config.GetConfig()
	fmt.Println("payload ->", *payload)
	req, err := http.NewRequest("POST", sapEndpoint, payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", config.GetString("SAP"))
	req.Header.Add("Content-Type", "application/json")

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
