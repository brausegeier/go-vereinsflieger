package vereinsflieger

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type Client struct {
	http.Client
}

func New() (c *Client, err error) {
	j, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	c = &Client{http.Client{Jar: j}}
	return
}

func (c *Client) Authenticate(user string, password string) (err error) {
	h := md5.Sum([]byte(password))
	l := url.Values{"user": {user}, "pwinput": {""}, "pw": {fmt.Sprintf("%x", h)}, "tan": {""}}
	_, err = c.PostForm("https://www.vereinsflieger.de/member/overview/index.php", l)
	return
}

func (c *Client) Logout() (err error) {
	_, err = c.Get("https://www.vereinsflieger.de/signout.php?signout=1")
	return
}

const voucherAddUrl = "https://www.vereinsflieger.de/member/community/addvoucher.php"

var voucherAddTKeyRegex = regexp.MustCompile("<input type='hidden' name='tkey' value='([^']*)'>")

func (c *Client) AddVoucher(v *Voucher, prefix string) (err error) {
	v.Identifier, err = c.nextVoucherIdentifier(prefix)
	if err != nil {
		return
	}
	resp, err := c.Get(voucherAddUrl)
	if err != nil {
		return
	}
	tKey, err := extractSubmatch(resp, voucherAddTKeyRegex)
	if err != nil {
		err = errors.New("Could not extract Transaction Key.")
		return
	}
	resp, err = c.PostForm(voucherAddUrl, *v.Values(tKey))
	m, err := regexp.MatchReader("class=\"message success", bufio.NewReader(resp.Body))
	if err != nil || !m {
		err = errors.New("Could not create voucher.")
	}
	return
}

const existingVoucherUrl = "https://www.vereinsflieger.de/member/community/voucher.php?sort=col1_desc"

func (c *Client) nextVoucherIdentifier(prefix string) (identifier string, err error) {
	sortFilterVals := url.Values{"col1": {string(prefix)}, "page": {"1"}, "submit": {"OK"}}
	resp, err := c.PostForm(existingVoucherUrl, sortFilterVals)
	if err != nil {
		return
	}
	year := time.Now().Year()
	regexStr := fmt.Sprintf("%s-%d-(\\d+)", prefix, year)
	fmt.Println("Looking for Vouchers matching: ", regexStr)
	rxp, err := regexp.Compile(regexStr)
	if err != nil {
		return
	}
	prevStr, err := extractSubmatch(resp, rxp)
	var next int = 1
	if err == nil {
		if prev, err := strconv.Atoi(prevStr); err == nil {
			next = prev + 1
		}
	}
	identifier = fmt.Sprintf("%s-%d-%03d", prefix, year, next)
	return
}

func extractSubmatch(response *http.Response, regex *regexp.Regexp) (match string, err error) {
	var b bytes.Buffer
	if _, err = io.Copy(&b, response.Body); err != nil {
		return
	}
	fmt.Println("Search in Text: ", b.String())
	n := regex.FindStringSubmatch(b.String())
	fmt.Println("Found vouchers: ", n)
	if len(n) < 2 {
		err = errors.New("Not Found.")
		return
	}
	match = n[1]
	return
}
