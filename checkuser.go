package main

func checkuser(phone string) bool {
	args := []string{phone}
	resp, err := cli.IsOnWhatsApp(args)
	if err != nil {
		wLog.Errorf("Failed to check if users are on WhatsApp:", err)
		return false
	}
	if len(resp) == 0 {
		wLog.Infof("No results")
		return false
	}

	item := resp[0]
	if item.VerifiedName != nil {
		wLog.Infof("%s: on whatsapp: %t, JID: %s, business name: %s", item.Query, item.IsIn, item.JID, item.VerifiedName.Details.GetVerifiedName())
	} else {
		wLog.Infof("%s: on whatsapp: %t, JID: %s", item.Query, item.IsIn, item.JID)
	}
	return item.IsIn
}
