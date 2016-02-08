package handler

import (
	"fmt"
	"github.com/janvogt/go-vereinsflieger/mailer"
	"github.com/janvogt/go-vereinsflieger/vereinsflieger"
	"github.com/janvogt/go-vereinsflieger/vereinsflieger/voucher"
	"github.com/naoina/toml"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Vereinsflieger struct {
		User     string
		Password string
	}
	Mail mailer.Mailer
}

type Gender int

const (
	Female Gender = 1
	Male   Gender = 2
)

var salutation = map[Gender]string{
	Female: "Frau",
	Male:   "Herr",
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
	http.HandleFunc("/voucher", Voucher)
}

func Voucher(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	var gender Gender
	switch req.PostForm.Get("buyer_gender") {
	case "1":
		gender = Female
	case "2":
		gender = Male
	default:
		http.Error(rw, "Invalid input for gender", http.StatusInternalServerError)
		return
	}
	var kind voucher.VoucherKind
	switch req.PostForm.Get("voucher_kind") {
	case "1":
		kind = voucher.Glider
	case "2":
		kind = voucher.MotorGlider
	default:
		http.Error(rw, "Invalid input for voucher kind", http.StatusInternalServerError)
		return
	}
	mins, err := strconv.ParseUint(req.PostForm.Get("duration"), 10, 32)
	if err != nil {
		http.Error(rw, "Invalid input for duration", http.StatusInternalServerError)
		return
	}
	v := voucher.Voucher{
		Title:         "Geschenkgutschein (Online, kein Versand)",
		Date:          time.Now(),
		Comment:       calcComment(req.PostForm.Get("buyer_firstname"), req.PostForm.Get("buyer_lastname"), gender, req.PostForm.Get("buyer_email")),
		Value:         calcValue(kind, uint(mins)),
		State:         voucher.Created,
		FirstName:     req.PostForm.Get("beneficiary_firstname"),
		LastName:      req.PostForm.Get("beneficiary_lastname"),
		Street:        req.PostForm.Get("beneficiary_street"),
		Zipcode:       req.PostForm.Get("beneficiary_zipcode"),
		City:          req.PostForm.Get("beneficiary_city"),
		Phone:         "",
		Mail:          "",
		CreateMemeber: false,
		Kind:          kind,
	}
	c, err := vereinsflieger.New()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("%+v\n", DefaultConfig)
	c.Authenticate(DefaultConfig.Vereinsflieger.User, DefaultConfig.Vereinsflieger.Password)
	defer c.Logout()
	err = v.Save(c)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	err = DefaultConfig.Mail.Voucher(req.PostForm.Get("buyer_email"), calcSalutation(gender, req.PostForm.Get("buyer_lastname")), v)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(rw, "Sucess")
}

func calcValue(k voucher.VoucherKind, mins uint) (value uint) {
	switch k {
	case voucher.Glider:
		value = 3500
	case voucher.MotorGlider:
		f := mins / 15
		if mins%15 > 0 {
			f += 1
		}
		if f < 2 {
			f = 2
		}
		value = 2250 * f
	}
	return
}

func calcComment(firstname, lastname string, gender Gender, email string) string {
	return fmt.Sprintf("Automatisch generiert für %s %s %s\nZahlungsaufforderung gesendet an %s", salutation[gender], firstname, lastname, email)
}

func calcSalutation(gender Gender, lastname string) string {
	var s string
	switch gender {
	case Female:
		s = "geehrte Frau"
	case Male:
		s = "geehrter Herr"
	}
	return fmt.Sprintf("Sehr %s %s,", s, lastname)
}
