package libs

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"github.com/fildenisov/go-whatsapp-rest/hlp"
	"math/rand"
	"os"
	"strings"
	"time"
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

func GetSendMutexSleepMS() time.Duration {
	rand.Seed(time.Now().UnixNano())
	min := 1000
	max := 3000
	return time.Duration(rand.Intn(max-min+1) + min)
}
