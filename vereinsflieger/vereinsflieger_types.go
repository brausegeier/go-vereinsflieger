package vereinsflieger

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type Gender int

const (
	Female Gender = 1
	Male          = 2
)

type Person struct {
	Gender    Gender
	GivenName string
	LastName  string
}

type GermanAdress struct {
	Street  string
	ZipCode string
	City    string
	Phone   string
	Mail    string
}

type Contact struct {
	Person
	GermanAdress
}

type VoucherState int

const (
	Created   VoucherState = 1
	Activated              = 2
	Used                   = 3
	Expired                = 4
)

type Voucher struct {
	Identifier    string
	Title         string
	Date          time.Time
	Comment       string
	Value         uint // value in cents
	State         VoucherState
	Beneficiary   Contact
	CreateMemeber bool
}

func (v Voucher) Values(tKey string) *url.Values {
	var vals = &url.Values{}
	vals.Add("frm_voucherid", v.Identifier)
	vals.Add("frm_title", v.Title)
	vals.Add("frm_voucherdate", v.Date.Format("02.01.2006"))
	vals.Add("frm_expiredate", "")
	vals.Add("frm_comment", v.Comment)
	vals.Add("frm_value", fmt.Sprintf("%d,%02d", v.Value/100, v.Value%100))
	vals.Add("frm_status", strconv.Itoa(int(v.State)))
	vals.Add("frm_firstname", v.Beneficiary.GivenName)
	vals.Add("frm_lastname", v.Beneficiary.LastName)
	vals.Add("frm_street", v.Beneficiary.Street)
	vals.Add("frm_zipcode", v.Beneficiary.ZipCode)
	vals.Add("frm_town", v.Beneficiary.City)
	vals.Add("frm_homenumber", v.Beneficiary.Phone)
	vals.Add("frm_email", v.Beneficiary.Mail)
	if v.CreateMemeber {
		vals.Add("frm_adduser", "1")
	}
	vals.Add("vid", "0")
	vals.Add("frm_uid", "0")
	vals.Add("tkey", tKey)
	return vals
}
