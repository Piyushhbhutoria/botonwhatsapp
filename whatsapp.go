package main

import (
	"encoding/gob"
	"fmt"
	"os"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qrChan := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qrChan).Print()
			// obj := qrcodeTerminal.New2(qrcodeTerminal.ConsoleColors.BrightBlue, qrcodeTerminal.ConsoleColors.BrightGreen, qrcodeTerminal.QRCodeRecoveryLevels.Low)
			// obj.Get(<-qrChan).Print()
		}()
		session, err = wac.Login(qrChan)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	if err = writeSession(session); err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(os.TempDir() + "/whatsapp.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&session); err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsapp.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = gob.NewEncoder(file).Encode(session); err != nil {
		return err
	}
	return nil
}
