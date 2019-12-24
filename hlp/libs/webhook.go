package libs

import (
	"bytes"
	"encoding/json"
	"github.com/dimaskiddo/go-whatsapp-rest/hlp"
	"net/http"
)

type HookRequest struct {
	Secret      string `json:"secret"`
	To          string `json:"to"`
	From        string `json:"from"`
	MessageType string `json:"message_type"`
	Message     string `json:"message"`
	FileName    string `json:"file_name"`
}

func HookData(jidFrom string, jidTo string, messageType string, message string, fileName string) error {
	req := &HookRequest{
		Secret:      hlp.Config.GetString("HOOK_SECRET"),
		To:          jidTo,
		From:        jidFrom,
		MessageType: messageType,
		Message:     message,
		FileName:    fileName,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	_, err = http.Post(hlp.Config.GetString("HOOK_URL"), "application/json", r)
	if err != nil {
		return err
	}
	return nil
}
