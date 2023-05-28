package main

import (
	"context"
	"net/http"
	"os"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func image(to, image, mess string) string {
	recipient, ok := parseJID(to)
	if !ok {
		return ""
	}
	check := checkuser(to)
	if check {
		wLog.Errorf("User doesn't exist: %v", to)
		return "User doesn't exist"
	}

	data, err := os.ReadFile(image)
	if err != nil {
		wLog.Errorf("Failed to read %s: %v", image, err)
		return "Error in reading image"
	}
	uploaded, err := cli.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		wLog.Errorf("Failed to upload file: %v", err)
		return "Error in uploading image"
	}
	msg := &waProto.Message{ImageMessage: &waProto.ImageMessage{
		Caption:       proto.String(to),
		Url:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(http.DetectContentType(data)),
		FileEncSha256: uploaded.FileEncSHA256,
		FileSha256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
	}}
	var respstr string
	ts, err := cli.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		wLog.Errorf("Error sending image message: %v", err)
		respstr = "Error sending image message: " + err.Error()
	} else {
		wLog.Infof("Image message sent (server timestamp: %s)", ts)
		respstr = "Image message Sent -> " + to + " : " + ts.ID
	}
	return respstr
}
