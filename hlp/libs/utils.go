package libs

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"github.com/dimaskiddo/go-whatsapp-rest/hlp"
	"os"
	"strings"
)

func GetMediaPath(info whatsapp.MessageInfo, rootFolder string, mediaType string) string {
	jid := ClearJid(info.RemoteJid)
	path := fmt.Sprintf("%v/%v/%v/%v", hlp.Config.GetString("SERVER_UPLOAD_PATH"), rootFolder, jid, mediaType)
	_ = os.MkdirAll(path, os.ModePerm)
	return path
}

func ClearJid(jid string) string {
	clearedJid := strings.Replace(jid, "@s.whatsapp.net", "", 1)
	clearedJid = strings.Replace(clearedJid, "@g.us", "", 1)
	return strings.Replace(clearedJid, "@c.us", "", 1)
}
