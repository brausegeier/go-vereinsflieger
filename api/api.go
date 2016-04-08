package api

import (
	"encoding/json"
	"fmt"
	"github.com/janvogt/go-vereinsflieger/mailer"
	"github.com/janvogt/go-vereinsflieger/vereinsflieger"
	"github.com/naoina/toml"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Port      uint
	Recaptcha struct {
		Secret  string
		Require bool
	}
	Vereinsflieger struct {
		User     string
		Password string
	}
	Mail mailer.Mailer
}

var DefaultConfig Config

func init() {
	b, err := ioutil.ReadFile(os.ExpandEnv("$GOPATH/src/github.com/janvogt/go-vereinsflieger/config.toml"))
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(b, &DefaultConfig)
	if err != nil {
		panic(err)
	}
}

func AddVoucher(rw http.ResponseWriter, req *http.Request) *HttpError {
	if err := req.ParseForm(); err != nil {
		return &HttpError{err, http.StatusBadRequest}
	}
	if !isHuman(req) && DefaultConfig.Recaptcha.Require {
		return &HttpError{fmt.Errorf("Unauthorized."), http.StatusForbidden}
	}
	v, err := NewVoucher(&req.PostForm)
	if err != nil {
		return &HttpError{err, http.StatusBadRequest}
	}
	c, err := vereinsflieger.New()
	if err != nil {
		return &HttpError{err, http.StatusInternalServerError}
	}
	c.Authenticate(DefaultConfig.Vereinsflieger.User, DefaultConfig.Vereinsflieger.Password)
	defer c.Logout()
	vv := v.ToVereinsflieger()
	if err := c.AddVoucher(v.ToVereinsflieger(), string(v.Kind)); err != nil {
		return &HttpError{err, http.StatusInternalServerError}
	}
	if err := DefaultConfig.Mail.Voucher(v.Client.Mail, calcSalutation(v), vv); err != nil {
		return &HttpError{err, http.StatusInternalServerError}
	}
	fmt.Fprintln(rw, "Success")
	return nil
}

func calcSalutation(v Voucher) string {
	var s string
	switch v.Client.Gender {
	case vereinsflieger.Female:
		s = "geehrte Frau"
	case vereinsflieger.Male:
		s = "geehrter Herr"
	}
	return fmt.Sprintf("Sehr %s %s,", s, v.Client.LastName)
}

const reCaptchaFormField = "g-recaptcha-response"
const reCaptchaVerifyUrl = "https://www.google.com/recaptcha/api/siteverify"

type ReCaptchaResult struct {
	Success   bool     `json:"success"`
	Timestamp string   `json:"challenge_ts"` // timestamp of the challenge load (ISO format yyyy-MM-dd'T'HH:mm:ssZZ)
	Hostname  string   `json:"hostname"`     // the hostname of the site where the reCAPTCHA was solved
	Errors    []string `json:"error-codes"`  // optional
}

// isHuman checks if an in-form reCaptcha proofs the request to be sent by a human.
func isHuman(req *http.Request) bool {
	response := req.PostForm.Get(reCaptchaFormField)
	if len(response) == 0 {
		return false
	}
	resp, err := http.PostForm(reCaptchaVerifyUrl, url.Values{"secret": {DefaultConfig.Recaptcha.Secret}, "response": {response}, "remoteip": {strings.Split(req.RemoteAddr, ":")[0]}})
	if err != nil {
		return false
	}
	var result ReCaptchaResult
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&result)
	if err != nil {
		return false
	}
	return result.Success
}
