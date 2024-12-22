package main

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	auth smtp.Auth
	addr string
	from string
}

func NewMailer(host string, port int, login string, passwd string, from string) *Mailer {
	return &Mailer{
		auth: smtp.PlainAuth("", login, passwd, host),
		addr: fmt.Sprintf("%s:%d", host, port),
		from: from,
	}
}

func (m *Mailer) SendMail(to string, text string) error {
	err := smtp.SendMail(m.addr, m.auth, m.from, []string{to}, []byte(text+"\r\n"))
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}
