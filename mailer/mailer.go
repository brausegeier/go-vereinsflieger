package mailer

import (
	"bytes"
	"fmt"
	"github.com/brausegeier/go-vereinsflieger/vereinsflieger"
	"net/smtp"
	"strings"
	"text/template"
)

type Mailer struct {
	VoucherTemplate string
	VoucherSubject  string
	User            string
	Password        string
	Host            string
	Port            int
	Sender          string
}

var SendMail = smtp.SendMail

type VoucherData struct {
	Salutation string
	Value      string
	Number     string
	Owner      string
}

func (m *Mailer) Voucher(to string, salutation string, v vereinsflieger.Voucher) (err error) {
	b := bytes.Buffer{}
	data := VoucherData{
		Salutation: salutation,
		Value:      fmt.Sprintf("%d,%02d €", v.Value/100, v.Value%100),
		Number:     v.Identifier,
		Owner:      fmt.Sprintf("%s, %s", v.Beneficiary.LastName, v.Beneficiary.GivenName),
	}
	t, err := template.New("VoucherTemplate").Parse(m.VoucherTemplate)
	if err != nil {
		return
	}
	err = t.Execute(&b, data)
	if err != nil {
		return
	}
	err = m.SendMail(to, m.VoucherSubject, b.String())
	if err != nil {
		fmt.Printf("Failed to send creation notification to: %s:\nSubject: %s\n\n%s\n", to, m.VoucherSubject, b.String())
	} else {
		fmt.Printf("Successfully notified: %s\nSubject: %s\n\n%s\n", to, m.VoucherSubject, b.String())
	}
	return
}

func (m *Mailer) SendMail(to string, subject string, body string) error {
	if !strings.ContainsRune(body, '\r') {
		body = strings.Replace(body, "\n", "\r\n", -1)
	}
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.Sender, to, subject, body))
	auth := smtp.PlainAuth("", m.User, m.Password, m.Host)
	return SendMail(fmt.Sprintf("%s:%d", m.Host, m.Port), auth, m.Sender, []string{to}, msg)
}
