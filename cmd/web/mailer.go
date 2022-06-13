package main

import (
	"bytes"
	"fmt"
	htmlTmpl "html/template"
	"sync"
	plainTmpl "text/template"

	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

// Define our mail server
type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	Encryption  string
	Wait        *sync.WaitGroup
	MailerChan  chan Message
	ErrorChan   chan error
	DoneChan    chan bool
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
	Template    string
}

func (app *Config) listenForMail() {
	for {
		select {
		case msg := <-app.Mailer.MailerChan:
			// spin off a mailer
			go app.Mailer.sendMail(msg, app.Mailer.ErrorChan)
		case err := <-app.Mailer.ErrorChan:
			// might want more here than just an error message!
			app.ErrorLog.Println(err)
		case <-app.Mailer.DoneChan:
			return // stop the goroutine I want to get off.
		}
	}
}

func (m *Mail) sendMail(msg Message, errorChan chan error) {
	defer m.Wait.Done()

	if msg.Template == "" {
		msg.Template = "mail"
	}

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}
	msg.DataMap = data

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		errorChan <- err
	}

	plainMessage, err := m.buildTextMessage(msg)
	if err != nil {
		errorChan <- err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		errorChan <- err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextPlain, plainMessage).
		AddAlternative(mail.TextHTML, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, atmt := range msg.Attachments {
			email.AddAttachment(atmt)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		errorChan <- err
	}

}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.html.gohtml", msg.Template)

	t, err := htmlTmpl.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	formattedMsg := tpl.String()

	html, err := m.inlineCSS(formattedMsg)
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) buildTextMessage(msg Message) (string, error) {
	if msg.Template == "" {
		msg.Template = "mail"
	}
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.plain.gohtml", msg.Template)

	t, err := plainTmpl.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.Encryption(mail.EncryptionNone)
	default:
		return mail.EncryptionSTARTTLS
	}
}
