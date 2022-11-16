package jobs

import (
	"context"
	"crypto/tls"

	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/scheduler"
	gomail "gopkg.in/mail.v2"
)

type MailEvent struct {
	To, Subject, ContentHTML string
}

var MailerJob = scheduler.NewJob("mailer_job", func(_ context.Context, ev MailEvent) error {
	m := gomail.NewMessage()
	c := config.Config()
	m.SetHeader("From", c.MailFrom)
	m.SetHeader("To", ev.To)
	m.SetHeader("Subject", ev.Subject)
	m.SetBody("text/html; charset=utf-8", ev.ContentHTML)

	username := c.SMTPUsername
	if username == "" {
		username = c.MailFrom
	}
	d := gomail.NewDialer(c.SMTPHost, c.SMTPPort, username, c.SMTPPassword)
	if !c.SMTPSecure {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return d.DialAndSend(m)
})
