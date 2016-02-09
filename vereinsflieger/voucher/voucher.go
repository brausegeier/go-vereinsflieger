package voucher

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VoucherState int

const (
	Created   VoucherState = 1
	Activated              = 2
	Used                   = 3
	Expired                = 4
)

type VoucherKind string

const (
	Glider      = "SF"
	MotorGlider = "TMG"
)

type Voucher struct {
	Number        string
	Title         string
	Date          time.Time
	Comment       string
	Value         uint // value in cents
	State         VoucherState
	FirstName     string
	LastName      string
	Street        string
	Zipcode       string
	City          string
	Phone         string
	Mail          string
	CreateMemeber bool
	Kind          VoucherKind
}

type StringFormPoster interface {
	PostFormString(string, url.Values) (string, error)
}

type StringGetter interface {
	GetString(string) (string, error)
}

type StringGetterFormPoster interface {
	StringFormPoster
	StringGetter
}

var debug bool = false

func (v *Voucher) nextNumber(c StringFormPoster) (result string, err error) {
	year := v.Date.Year()
	resp, err := c.PostFormString("https://www.vereinsflieger.de/member/community/voucher.php?sort=col1_desc", url.Values{"col1": {string(v.Kind)}, "page": {"1"}, "submit": {"OK"}})
	if err != nil {
		return
	}
	r, err := regexp.Compile(fmt.Sprintf("%s-%d-(\\d+)", v.Kind, year))
	if err != nil {
		return
	}
	n := r.FindStringSubmatch(resp)
	var next int = 1
	if len(n) > 1 {
		next, err = strconv.Atoi(n[1])
		if err != nil {
			return
		}
		next += 1
	}
	result = fmt.Sprintf("%s-%d-%03d", v.Kind, year, next)
	return
}

// frm_voucherid=SF-2016-002&frm_title=test&frm_voucherdate=06.02.2016&frm_comment=&frm_value=443%2C00&frm_status=1&frm_firstname=53w&frm_lastname=23&frm_street=&frm_zipcode=&frm_town=&frm_homenumber=&frm_email=&vid=2843&frm_uid=0&tkey=4239ed60e4198dc6bdd201341ebd583c

func (v *Voucher) Save(c StringGetterFormPoster) (err error) {
	r, err := c.GetString("https://www.vereinsflieger.de/member/community/addvoucher.php")
	if err != nil {
		return
	}
	rxp, err := regexp.Compile("<input type='hidden' name='tkey' value='([^']*)'>")
	if err != nil {
		return
	}
	k := rxp.FindStringSubmatch(r)
	if len(k) != 2 {
		err = errors.New("Could not extract Transaction Key.")
		return
	}
	vals := url.Values{}
	if v.Number == "" {
		(*v).Number, err = v.nextNumber(c)
		if err != nil {
			return
		}
	}
	vals.Add("frm_voucherid", v.Number)
	vals.Add("frm_title", v.Title)
	vals.Add("frm_voucherdate", v.Date.Format("02.01.2006"))
	vals.Add("frm_comment", v.Comment)
	vals.Add("frm_value", fmt.Sprintf("%d,%02d", v.Value/100, v.Value%100))
	vals.Add("frm_status", strconv.Itoa(int(v.State)))
	vals.Add("frm_firstname", v.FirstName)
	vals.Add("frm_lastname", v.LastName)
	vals.Add("frm_street", v.Street)
	vals.Add("frm_zipcode", v.Zipcode)
	vals.Add("frm_town", v.City)
	vals.Add("frm_homenumber", v.Phone)
	vals.Add("frm_email", v.Mail)
	vals.Add("vid", "0")
	vals.Add("frm_uid", "0")
	vals.Add("tkey", k[1])
	r, err = c.PostFormString("https://www.vereinsflieger.de/member/community/addvoucher.php", vals)
	if !strings.Contains(r, "class=\"message success") {
		err = errors.New("Could not create voucher.")
	}
	return
}
