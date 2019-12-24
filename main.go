package main

import (
	"fmt"
	"github.com/fildenisov/go-whatsapp-rest/ctl"
	"github.com/fildenisov/go-whatsapp-rest/hlp/auth"
	"os"
	"os/signal"
	"syscall"

	"github.com/dimaskiddo/go-whatsapp-rest/hlp"
	"github.com/dimaskiddo/go-whatsapp-rest/hlp/router"
)

// Server Variable
var svr *hlp.Server

// Init Function
func init() {
	// Initialize Server
	svr = hlp.NewServer(router.Router)

	// Set Endpoint for Root Functions
	router.Router.Get(router.RouterBasePath, ctl.GetIndex)
	router.Router.Get(router.RouterBasePath+"/health", ctl.GetHealth)

	// Set Endpoint for Authorization Functions
	router.Router.With(auth.Basic).Get(router.RouterBasePath+"/auth", ctl.GetAuth)

	// Set Endpoint for WhatsApp Functions
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/login", ctl.WhatsAppLogin)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/text", ctl.WhatsAppSendText)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/image", ctl.WhatsAppSendImage)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/location", ctl.WhatsAppSendLocation)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/document", ctl.WhatsAppSendDocument)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/video", ctl.WhatsAppSendVideo)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/logout", ctl.WhatsAppLogout)
	router.Router.Get(router.RouterBasePath+"/files/*", ctl.GetFile)

	ctl.ConnectAllSessions()
}

// Main Function
func main() {
	// Starting Server
	svr.Start()

	// Make Channel for OS Signal
	sig := make(chan os.Signal, 1)

	// Notify Any Signal to OS Signal Channel
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	// Return OS Signal Channel
	// As Exit Sign
	<-sig

	// Log Break Line
	fmt.Println("")

	// Stopping Server
	defer svr.Stop()
}
