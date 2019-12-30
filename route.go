package main

import (
	"github.com/fildenisov/go-whatsapp-rest/ctl"
	"github.com/fildenisov/go-whatsapp-rest/hlp/auth"
	"github.com/fildenisov/go-whatsapp-rest/hlp/router"
)

// Initialize Function in Main Route
func init() {
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
