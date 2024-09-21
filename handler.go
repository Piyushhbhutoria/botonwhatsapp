package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	historySyncID int32
	startupTime   = time.Now().Unix()
)

func parseJID(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	}
	recipient, err := types.ParseJID(arg)
	if err != nil {
		wLog.Errorf("Invalid JID %s: %v", arg, err)
		return recipient, false
	} else if recipient.User == "" {
		wLog.Errorf("Invalid JID %s: no server specified", arg)
		return recipient, false
	}
	return recipient, true
}

func handler(rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.AppStateSyncComplete:
		if len(cli.Store.PushName) > 0 && evt.Name == appstate.WAPatchCriticalBlock {
			err := cli.SendPresence(types.PresenceAvailable)
			if err != nil {
				wLog.Warnf("Failed to send available presence: %v", err)
			} else {
				wLog.Infof("Marked self as available")
			}
		}
	case *events.Connected, *events.PushNameSetting:
		if len(cli.Store.PushName) == 0 {
			return
		}
		// Send presence available when connecting and when the pushname is changed.
		// This makes sure that outgoing messages always have the right pushname.
		err := cli.SendPresence(types.PresenceAvailable)
		if err != nil {
			wLog.Warnf("Failed to send available presence: %v", err)
		} else {
			wLog.Infof("Marked self as available")
		}
	case *events.StreamReplaced:
		os.Exit(0)
	case *events.Message:
		message := evt.Message.GetConversation()
		if evt.Info.Timestamp.Unix() > startupTime {
			metaParts := []string{fmt.Sprintf("pushname: %s", evt.Info.PushName), fmt.Sprintf("timestamp: %s", evt.Info.Timestamp)}
			if evt.Info.Type != "" {
				metaParts = append(metaParts, fmt.Sprintf("type: %s", evt.Info.Type))
			}
			if evt.Info.Category != "" {
				metaParts = append(metaParts, fmt.Sprintf("category: %s", evt.Info.Category))
			}
			if evt.IsViewOnce {
				metaParts = append(metaParts, "view once")
			}
			if evt.IsViewOnce {
				metaParts = append(metaParts, "ephemeral")
			}

			wLog.Infof("Received message %s from %s (%s): %s", evt.Info.ID, evt.Info.SourceString(), strings.Join(metaParts, ", "), message)

			if evt.Info.Timestamp.After(now) && !evt.Info.IsFromMe && !evt.Info.IsGroup && len(message) < 21 { // TODO: update condition
				requestChannel <- evt
			}

			if !strings.Contains(message, "status@broadcast") {
				img := evt.Message.GetImageMessage()
				if img != nil {
					data, err := cli.Download(img)
					if err != nil {
						wLog.Errorf("Failed to download image: %v", err)
						return
					}
					exts, _ := mime.ExtensionsByType(img.GetMimetype())
					path := fmt.Sprintf("/images/%s%s", evt.Info.ID, exts[0])
					err = os.WriteFile(path, data, 0600)
					if err != nil {
						wLog.Errorf("Failed to save image: %v", err)
						return
					}
					wLog.Infof("Saved image in message to %s", path)
				}
			}
		}
	case *events.Receipt:
		if evt.Type == types.ReceiptTypeRead || evt.Type == types.ReceiptTypeReadSelf {
			wLog.Infof("%v was read by %s at %s", evt.MessageIDs, evt.SourceString(), evt.Timestamp)
		} else if evt.Type == types.ReceiptTypeDelivered {
			wLog.Infof("%s was delivered to %s at %s", evt.MessageIDs[0], evt.SourceString(), evt.Timestamp)
		}
	case *events.Presence:
		if evt.Unavailable {
			if evt.LastSeen.IsZero() {
				wLog.Infof("%s is now offline", evt.From)
			} else {
				wLog.Infof("%s is now offline (last seen: %s)", evt.From, evt.LastSeen)
			}
		} else {
			wLog.Infof("%s is now online", evt.From)
		}
	case *events.HistorySync:
		id := atomic.AddInt32(&historySyncID, 1)
		fileName := fmt.Sprintf("history-%d-%d.json", startupTime, id)
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			wLog.Errorf("Failed to open file to write history sync: %v", err)
			return
		}
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		err = enc.Encode(evt.Data)
		if err != nil {
			wLog.Errorf("Failed to write history sync: %v", err)
			return
		}
		wLog.Infof("Wrote history sync to %s", fileName)
		_ = file.Close()
	case *events.AppState:
		wLog.Debugf("App state event: %+v / %+v", evt.Index, evt.SyncActionValue)
	}
}
