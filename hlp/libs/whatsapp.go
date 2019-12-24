package libs

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/Rhymen/go-whatsapp"
	waproto "github.com/Rhymen/go-whatsapp/binary/proto"
	"github.com/skip2/go-qrcode"

	"github.com/dimaskiddo/go-whatsapp-rest/hlp"
)

type waHandler struct {
	c *whatsapp.Conn
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (this *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.FromMe || hlp.Config.GetString("HOOK_URL") == "" {
		return
	}

	_ = HookData(
		ClearJid(message.Info.RemoteJid),
		ClearJid(this.c.Info.Wid),
		"text",
		message.Text,
		"",
	)
}

func (this *waHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	if message.Info.FromMe || hlp.Config.GetString("HOOK_URL") == "" {
		return
	}

	imageData, err := message.Download()
	if err != nil {
		fmt.Println(err)
		return
	}

	jidTo := ClearJid(this.c.Info.Wid)
	path := GetMediaPath(message.Info, jidTo, "images")
	fileName := fmt.Sprintf("%v/%v.jpg", path, message.Info.Id)
	err = ioutil.WriteFile(fileName, imageData, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = HookData(
		ClearJid(message.Info.RemoteJid),
		ClearJid(this.c.Info.Wid),
		"image",
		message.Caption,
		message.Info.Id+".jpg",
	)
}

func (this *waHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	if message.Info.FromMe || hlp.Config.GetString("HOOK_URL") == "" {
		return
	}

	imageData, err := message.Download()
	if err != nil {
		fmt.Println(err)
		return
	}

	path := GetMediaPath(message.Info, ClearJid(this.c.Info.Wid), "documents")
	fileName := fmt.Sprintf("%v/%v", path, message.FileName)
	err = ioutil.WriteFile(fileName, imageData, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = HookData(
		ClearJid(message.Info.RemoteJid),
		ClearJid(this.c.Info.Wid),
		"document",
		message.Title,
		message.FileName,
	)
}

func (this *waHandler) HandleVideoMessage(message whatsapp.VideoMessage) {
	if message.Info.FromMe || hlp.Config.GetString("HOOK_URL") == "" {
		return
	}

	imageData, err := message.Download()
	if err != nil {
		fmt.Println(err)
		return
	}

	path := GetMediaPath(message.Info, ClearJid(this.c.Info.Wid), "videos")
	fileName := fmt.Sprintf("%v/%v.mp4", path, message.Info.Id)
	err = ioutil.WriteFile(fileName, imageData, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = HookData(
		ClearJid(message.Info.RemoteJid),
		ClearJid(this.c.Info.Wid),
		"video",
		message.Caption,
		message.Info.Id+".mp4",
	)
}

func (this *waHandler) HandleLocationMessage(message whatsapp.LocationMessage) {
	if message.Info.FromMe || hlp.Config.GetString("HOOK_URL") == "" {
		return
	}

	_ = HookData(
		ClearJid(message.Info.RemoteJid),
		ClearJid(this.c.Info.Wid),
		"location",
		fmt.Sprintf("%v,%v", message.DegreesLatitude, message.DegreesLongitude),
		"",
	)
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

var wac = make(map[string]*whatsapp.Conn)

func WASyncVersion(conn *whatsapp.Conn) (string, error) {
	versionServer, err := whatsapp.CheckCurrentServerVersion()
	if err != nil {
		return "", err
	}

	conn.SetClientVersion(versionServer[0], versionServer[1], versionServer[2])
	versionClient := conn.GetClientVersion()

	return fmt.Sprintf("whatsapp version %v.%v.%v", versionClient[0], versionClient[1], versionClient[2]), nil
}

func WATestPing(conn *whatsapp.Conn) error {
	ok, err := conn.AdminTest()
	if !ok {
		if err != nil {
			return err
		} else {
			return errors.New("something when wrong while trying to ping, please check phone connectivity")
		}
	}

	return nil
}

func WAGenerateQR(timeout int, chanqr chan string, qrstr chan<- string) {
	select {
	case tmp := <-chanqr:
		png, _ := qrcode.Encode(tmp, qrcode.Medium, 256)
		qrstr <- base64.StdEncoding.EncodeToString(png)
	}
}

func WASessionInit(jid string, timeout int) error {
	if wac[jid] == nil {
		conn, err := whatsapp.NewConn(time.Duration(timeout) * time.Second)
		if err != nil {
			return err
		}

		conn.SetClientName(hlp.Config.GetString("LONG_CLIENT_NAME"), hlp.Config.GetString("SHORT_CLIENT_NAME"))

		info, err := WASyncVersion(conn)
		if err != nil {
			return err
		}
		hlp.LogPrintln(hlp.LogLevelInfo, "whatsapp", info)

		wac[jid] = conn
		go WAAddHandlers(jid)
	}

	return nil
}

func WASessionLoad(file string) (whatsapp.Session, error) {
	session := whatsapp.Session{}

	buffer, err := os.Open(file)
	if err != nil {
		return session, err
	}
	defer buffer.Close()

	err = gob.NewDecoder(buffer).Decode(&session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func WASessionSave(file string, session whatsapp.Session) error {
	buffer, err := os.Create(file)
	if err != nil {
		return err
	}
	defer buffer.Close()

	err = gob.NewEncoder(buffer).Encode(session)
	if err != nil {
		return err
	}

	return nil
}

func WASessionExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}

	return true
}

func WASessionConnect(jid string, timeout int, file string, qrstr chan<- string, errmsg chan<- error) {
	chanqr := make(chan string)

	session, err := WASessionLoad(file)
	if err != nil {
		go func() {
			WAGenerateQR(timeout, chanqr, qrstr)
		}()

		err = WASessionLogin(jid, timeout, file, chanqr)
		if err != nil {
			errmsg <- err
			return
		}
		return
	}

	err = WASessionRestore(jid, timeout, file, session)
	if err != nil {
		go func() {
			WAGenerateQR(timeout, chanqr, qrstr)
		}()

		err = WASessionLogin(jid, timeout, file, chanqr)
		if err != nil {
			errmsg <- err
			return
		}
	}

	err = WATestPing(wac[jid])
	if err != nil {
		errmsg <- err
		return
	}

	errmsg <- errors.New("")
	return
}

func WASessionLogin(jid string, timeout int, file string, qrstr chan<- string) error {
	if wac[jid] != nil {
		if WASessionExist(file) {
			err := os.Remove(file)
			if err != nil {
				return err
			}
		}

		delete(wac, jid)
	}

	err := WASessionInit(jid, timeout)
	if err != nil {
		return err
	}

	session, err := wac[jid].Login(qrstr)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "already logged in":
			return nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return errors.New("connection is invalid")
		default:
			delete(wac, jid)
			return err
		}
	}

	err = WASessionSave(file, session)
	if err != nil {
		return err
	}

	return nil
}

func WASessionRestore(jid string, timeout int, file string, sess whatsapp.Session) error {
	if wac[jid] != nil {
		if WASessionExist(file) {
			err := os.Remove(file)
			if err != nil {
				return err
			}
		}

		delete(wac, jid)
	}

	err := WASessionInit(jid, timeout)
	if err != nil {
		return err
	}

	session, err := wac[jid].RestoreWithSession(sess)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "already logged in":
			return nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return errors.New("connection is invalid")
		default:
			delete(wac, jid)
			return err
		}
	}

	err = WASessionSave(file, session)
	if err != nil {
		return err
	}

	return nil
}

func WASessionLogout(jid string, file string) error {
	if wac[jid] != nil {
		err := wac[jid].Logout()
		if err != nil {
			return err
		}

		if WASessionExist(file) {
			err = os.Remove(file)
			if err != nil {
				return err
			}
		}

		delete(wac, jid)
	} else {
		return errors.New("connection is invalid")
	}

	return nil
}

func WAMessageText(jid string, jidDest string, msgText string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Text: msgText,
		}

		if len(msgQuotedID) != 0 {
			pntQuotedMsg := &waproto.Message{
				Conversation: &msgQuoted,
			}

			content.Info.QuotedMessageID = msgQuotedID
			content.Info.QuotedMessage = *pntQuotedMsg
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return "", errors.New("connection is invalid")
			default:
				return "", err
			}
		}
	} else {
		return "", errors.New("connection is invalid")
	}

	return id, nil
}

func WAMessageLocation(jid string, jidDest string, degreesLatitude float64, degreesLongitude float64, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.LocationMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			DegreesLatitude:  degreesLatitude,
			DegreesLongitude: degreesLongitude,
		}

		if len(msgQuotedID) != 0 {
			pntQuotedMsg := &waproto.Message{
				Conversation: &msgQuoted,
			}

			content.Info.QuotedMessageID = msgQuotedID
			content.Info.QuotedMessage = *pntQuotedMsg
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return "", errors.New("connection is invalid")
			default:
				return "", err
			}
		}
	} else {
		return "", errors.New("connection is invalid")
	}

	return id, nil
}

func WAMessageImage(jid string, jidDest string, msgImageStream multipart.File, msgImageType string, msgCaption string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.ImageMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Content: msgImageStream,
			Type:    msgImageType,
			Caption: msgCaption,
		}

		if len(msgQuotedID) != 0 {
			pntQuotedMsg := &waproto.Message{
				Conversation: &msgQuoted,
			}

			content.Info.QuotedMessageID = msgQuotedID
			content.Info.QuotedMessage = *pntQuotedMsg
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return "", errors.New("connection is invalid")
			default:
				return "", err
			}
		}
	} else {
		return "", errors.New("connection is invalid")
	}

	return id, nil
}

func WAMessageVideo(jid string, jidDest string, msgVideoStream multipart.File, msgVideoType string, msgCaption string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.VideoMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Content: msgVideoStream,
			Type:    msgVideoType,
			Caption: msgCaption,
		}

		if len(msgQuotedID) != 0 {
			pntQuotedMsg := &waproto.Message{
				Conversation: &msgQuoted,
			}

			content.Info.QuotedMessageID = msgQuotedID
			content.Info.QuotedMessage = *pntQuotedMsg
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return "", errors.New("connection is invalid")
			default:
				return "", err
			}
		}
	} else {
		return "", errors.New("connection is invalid")
	}

	return id, nil
}

func WAMessageDocument(jid string, jidDest string, msgDocumentStream multipart.File, msgDocumentType string, msgDocumentFileName string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.DocumentMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Content:  msgDocumentStream,
			Type:     msgDocumentType,
			FileName: msgDocumentFileName,
			Title:    msgDocumentFileName,
		}

		if len(msgQuotedID) != 0 {
			pntQuotedMsg := &waproto.Message{
				Conversation: &msgQuoted,
			}

			content.Info.QuotedMessageID = msgQuotedID
			content.Info.QuotedMessage = *pntQuotedMsg
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return "", errors.New("connection is invalid")
			default:
				return "", err
			}
		}
	} else {
		return "", errors.New("connection is invalid")
	}

	return id, nil
}

func WAAddHandlers(jid string) {
	//sleep 10 sec to not handle old messages
	time.Sleep(time.Duration(10) * time.Second)
	hlp.LogPrintln(hlp.LogLevelInfo, "handlers", "handlers for  "+jid+" added")
	wac[jid].AddHandler(&waHandler{wac[jid]})
}
