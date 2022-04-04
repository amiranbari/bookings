package main

import (
	"github.com/amiranbari/bookings/pkg/models"
	mail "github.com/xhit/go-simple-mail/v2"
	"time"
)

const mailTimeOut = 10

func listenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			sendMail(msg)
		}
	}()
}

func sendMail(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 11025
	server.KeepAlive = false
	server.ConnectTimeout = mailTimeOut * time.Second
	server.SendTimeout = mailTimeOut * time.Second

	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject).SetBody(mail.TextHTML, m.Content)

	err = email.Send(client)
	if err != nil {
		errorLog.Println(err)
	} else {
		infoLog.Println("Email sent.")
	}

}
