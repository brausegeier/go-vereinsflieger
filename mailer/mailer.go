package mailer

import (
	"bytes"
	"fmt"
	"github.com/janvogt/go-vereinsflieger/vereinsflieger/voucher"
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

func (m *Mailer) Voucher(to string, salutation string, v voucher.Voucher) (err error) {
	b := bytes.Buffer{}
	data := VoucherData{
		Salutation: salutation,
		Value:      fmt.Sprintf("%d,%02d €", v.Value/100, v.Value%100),
		Number:     v.Number,
		Owner:      fmt.Sprintf("%s, %s", v.LastName, v.FirstName),
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
