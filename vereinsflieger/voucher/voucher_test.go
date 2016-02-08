package voucher

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

type ValueResult struct {
	vals   url.Values
	result string
}

type Mapping map[string]ValueResult

type MapSFP struct {
	mapping Mapping
	test    *testing.T
	calls   int
}

func (s *MapSFP) PostFormString(url string, vals url.Values) (string, error) {
	v, ok := s.mapping[url]
	if !ok {
		s.test.Errorf("Form Post to unexpected URL %s", url)
	}
	if v.vals != nil && !reflect.DeepEqual(vals, v.vals) {
		s.test.Errorf("%+s != %+s", vals, v.vals)
	}
	s.calls += 1
	return v.result, nil
}

func (s *MapSFP) GetString(url string) (string, error) {
	v, ok := s.mapping[url]
	if !ok {
		s.test.Errorf("Get to unexpected URL %s", url)
	}
	s.calls += 1
	return v.result, nil
}

const NextUrl = "https://www.vereinsflieger.de/member/community/voucher.php?sort=col1_desc"

var NextTests = []struct {
	kind   VoucherKind
	con    MapSFP
	result string
}{
	{Glider, MapSFP{Mapping{NextUrl: ValueResult{url.Values{"col1": {"SF"}, "page": {"1"}, "submit": {"OK"}}, "<many><html>SF-2016-003,TMG-2016-002,SF-2016-001</tadaa>"}}, nil, 0}, "SF-2016-004"},
	{MotorGlider, MapSFP{Mapping{NextUrl: ValueResult{url.Values{"col1": {"TMG"}, "page": {"1"}, "submit": {"OK"}}, "<many><html>SF-2016-003,TMG-2016-002,SF-2016-001</tadaa>"}}, nil, 0}, "TMG-2016-003"},
	{Glider, MapSFP{Mapping{NextUrl: ValueResult{url.Values{"col1": {"SF"}, "page": {"1"}, "submit": {"OK"}}, "<many><html>=TMG-2016-002,6-001</tadaa>"}}, nil, 0}, "SF-2016-001"},
}

func TestNext(t *testing.T) {
	year2016, _ := time.Parse("2006", "2016")
	for _, tt := range NextTests {
		v := Voucher{Kind: tt.kind, Date: year2016}
		tt.con.test = t
		result, err := v.nextNumber(&tt.con)
		if err != nil {
			t.Fatal(err)
		}
		if result != tt.result {
			t.Errorf("%s != %s", result, tt.result)
		}
	}
	return
}

const AddVoucherUrl = "https://www.vereinsflieger.de/member/community/addvoucher.php"
const VoucherURL = "https://www.vereinsflieger.de/member/community/voucher.php"

var SaveTests = []struct {
	voucher Voucher
	con     MapSFP
}{
	{
		Voucher{"number", "title", time.Date(2016, 2, 6, 0, 0, 0, 0, time.UTC), "comment", 12345, Created, "firstname", "lastname", "street", "zipcode", "city", "phone", "mail", false, Glider},
		MapSFP{
			Mapping{
				AddVoucherUrl: ValueResult{url.Values{"frm_voucherid": {"number"}, "frm_title": {"title"}, "frm_voucherdate": {"06.02.2016"}, "frm_comment": {"comment"}, "frm_value": {"123,45"}, "frm_status": {"1"}, "frm_firstname": {"firstname"}, "frm_lastname": {"lastname"}, "frm_street": {"street"}, "frm_zipcode": {"zipcode"}, "frm_town": {"city"}, "frm_homenumber": {"phone"}, "frm_email": {"mail"}, "vid": {"0"}, "frm_uid": {"0"}, "tkey": {"transactionkey"}}, "<many><html><input type='hidden' name='tkey' value='transactionkey'>value=\"transactionkey\"></tadaa><div id='messagecontainer'><div id=\"msg_23989961\" class=\"message success ...\">sdfds"},
			},
			nil,
			0,
		},
	},
}

func TestSave(t *testing.T) {
	for _, tt := range SaveTests {
		tt.con.test = t
		err := tt.voucher.Save(&tt.con)
		if err != nil {
			t.Fatal(err)
		}
		if tt.con.calls < 2 {
			t.Error("Not all expected calls have been made.")
		}
	}
	return
}
