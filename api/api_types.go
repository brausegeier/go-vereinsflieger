package api

import (
	"fmt"
	"github.com/brausegeier/go-vereinsflieger/vereinsflieger"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type VoucherKind string

const (
	Glider      = "SF"
	MotorGlider = "TMG"
)

type ClientContact struct {
	vereinsflieger.Person
	Mail string
}

type Voucher struct {
	Kind        VoucherKind
	Duration    uint // Duration in minutes
	Client      ClientContact
	Beneficiary vereinsflieger.Contact
}

func NewVoucher(values *url.Values) (v Voucher, err error) {
	switch values.Get("voucher_kind") {
	case "1":
		v.Kind = Glider
	case "2":
		v.Kind = MotorGlider
	default:
		err = fmt.Errorf("Invalid Value %s for voucher_kind", values.Get("voucher_kind"))
		return
	}
	duration, err := strconv.ParseUint(values.Get("duration"), 10, strconv.IntSize)
	if err != nil && v.Kind == MotorGlider {
		err = fmt.Errorf("Invalid Value %s for duration", values.Get("duration"))
		return
	} else {
		err = nil
	}
	v.Duration = uint(duration)
	switch values.Get("buyer_gender") {
	case "1":
		v.Client.Gender = vereinsflieger.Female
	case "2":
		v.Client.Gender = vereinsflieger.Male
	default:
		err = fmt.Errorf("Invalid Value %s for buyer_gender", values.Get("buyer_gender"))
		return
	}
	v.Client.GivenName = values.Get("buyer_firstname")
	v.Client.LastName = values.Get("buyer_lastname")
	v.Client.Mail = values.Get("buyer_email")
	v.Beneficiary.GivenName = values.Get("beneficiary_firstname")
	v.Beneficiary.LastName = values.Get("beneficiary_lastname")
	v.Beneficiary.Street = values.Get("beneficiary_street")
	v.Beneficiary.ZipCode = values.Get("beneficiary_zipcode")
	v.Beneficiary.City = values.Get("beneficiary_city")
	return
}

var salutation = map[vereinsflieger.Gender]string{
	vereinsflieger.Female: "Frau",
	vereinsflieger.Male:   "Herr",
}

func calcValue(k VoucherKind, mins uint) (value uint) {
	switch k {
	case Glider:
		value = 3500
	case MotorGlider:
		f := mins / 15
		if mins%15 > 0 {
			f += 1
		}
		if f < 2 {
			f = 2
		}
		value = 2750 * f
	}
	return
}

func (v *Voucher) ToVereinsflieger() (vv vereinsflieger.Voucher) {
	vv.Beneficiary = v.Beneficiary
	vv.Comment = fmt.Sprintf("Automatisch generiert für %s %s %s\nZahlungsaufforderung gesendet an %s", salutation[v.Client.Gender], v.Client.GivenName, v.Client.LastName, v.Client.Mail)
	vv.CreateMemeber = false
	vv.Date = time.Now()
	vv.State = vereinsflieger.Created
	vv.Title = "Geschenkgutschein (Online, kein Versand)"
	vv.Value = calcValue(v.Kind, v.Duration)
	return
}

type HttpError struct {
	error
	Status   int
	Redirect string
}

type FailableHandler func(http.ResponseWriter, *http.Request) *HttpError

func (fh FailableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := fh(rw, req); err != nil {
		if err.Redirect != "" {
			http.Redirect(rw, req, err.Redirect, http.StatusSeeOther)
		} else {
			http.Error(rw, err.Error(), err.Status)
		}
	}
}
