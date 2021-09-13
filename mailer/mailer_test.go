package mailer

import (
	"github.com/brausegeier/go-vereinsflieger/vereinsflieger"
	"net/smtp"
	"strings"
	"testing"
)

type SendMailMock struct {
	addr   string
	auth   smtp.Auth
	sender string
	to     []string
	msg    []byte
	res    error
}

func (s *SendMailMock) SendMail(addr string, auth smtp.Auth, sender string, to []string, msg []byte) error {
	s.addr = addr
	s.auth = auth
	s.sender = sender
	s.to = to
	s.msg = msg
	return s.res
}

func TestVoucher(t *testing.T) {
	s := SendMailMock{}
	SendMail = s.SendMail
	m := Mailer{
		User:            "user",
		Password:        "pw",
		Host:            "host",
		Port:            25,
		Sender:          "sender",
		VoucherTemplate: vTemplate,
		VoucherSubject:  "subject",
	}
	m.Voucher("to", "salutation", vereinsflieger.Voucher{
		Identifier: "number",
		Value:      12300,
		Beneficiary: vereinsflieger.Contact{
			Person: vereinsflieger.Person{
				GivenName: "firstname",
				LastName:  "lastname",
			},
		},
	})
	if !strings.Contains(string(s.msg), vResult) {
		t.Errorf("Unexpected Body: \n%v", s.msg)
		t.Errorf("Unexpected Body: \n%v", []byte(vResult))
	}
}

const vTemplate = `
Hallo {{.Salutation}},

Gutschein über Betrag {{.Value}}: angelegt: {{.Number}} {{.Owner}} angelegt.

Bitte Zahlen!
`

const vResult = "Hallo salutation,\r\n\r\nGutschein über Betrag 123,00 €: angelegt: number lastname, firstname angelegt.\r\n\r\nBitte Zahlen!\r\n"
