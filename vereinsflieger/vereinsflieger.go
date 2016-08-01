package vereinsflieger

import (
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

const loginUrl = "https://www.vereinsflieger.de/"

var pwsaltRegex = regexp.MustCompile("<input type=\"hidden\" name=\"pwdsalt\" value=\"([^']*)\" />")

func (c *Client) Authenticate(user string, password string) (err error) {
	h := md5.Sum([]byte(password))
	resp, err := c.Get(loginUrl)
	if err != nil {
		return
	}
	fmt.Println("Looking for Password Salt")
	salt, _, err := extractSubmatch(resp, pwsaltRegex)
	if err != nil {
		err = errors.New("Could not extract pwdsalt.")
		return
	}
	l := url.Values{"user": {user}, "pwinput": {""}, "pw": {fmt.Sprintf("%x", h)}, "tan": {""}, "pwdsalt": {salt}}
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

	fmt.Println("Looking for Transaction Key")
	tKey, _, err := extractSubmatch(resp, voucherAddTKeyRegex)
	if err != nil {
		err = errors.New("Could not extract Transaction Key.")
		return
	}
	resp, err = c.PostForm(voucherAddUrl, *v.Values(tKey))
	respBuf := bytes.Buffer{}
	_, err = respBuf.ReadFrom(resp.Body)
	if err != nil {
		fmt.Println("No body when reading creation response.")
		err = errors.New("Could not create voucher.")
	}
	m, err := regexp.MatchReader("class=\"message success", bytes.NewReader(respBuf.Bytes()))
	if err != nil || !m {
		fmt.Println("No success message. Body was:", respBuf.String())
		err = errors.New("Could not create voucher.")
	} else {
		fmt.Printf("Created Voucher %+v", v)
	}
	return
}

const existingVoucherUrl = "https://www.vereinsflieger.de/member/community/voucher.php?sort=col1_desc"

func (c *Client) nextVoucherIdentifier(prefix string) (identifier string, err error) {
	year := time.Now().Year()
	next, err := c.nextVoucherIndexFromPage(prefix, year, 1)
	if err == nil {
		identifier = fmt.Sprintf("%s-%d-%03d", prefix, year, next)
	}
	return
}

func (c *Client) nextVoucherIndexFromPage(prefix string, year int, page int) (index int, err error) {
	sortFilterVals := url.Values{"col1": {string(prefix)}, "page": {strconv.Itoa(page)}, "submit": {"OK"}}
	resp, err := c.PostForm(existingVoucherUrl, sortFilterVals)
	if err != nil {
		return
	}
	regexStr := fmt.Sprintf("%s-%d-(\\d+)", prefix, year)
	fmt.Printf("Looking for Vouchers matching on page %d: %v", page, regexStr)
	rxp, err := regexp.Compile(regexStr)
	if err != nil {
		return
	}
	prevStrings, result, err := extractSubmatches(resp, rxp)
	index = 1
	for _, prevStr := range prevStrings {
		if prev, err := strconv.Atoi(prevStr); err == nil {
			if index < prev+1 {
				index = prev + 1
			}
		}
	}
	if c.hasNextPage(result) {
		var nextPageIndex int
		nextPageIndex, err = c.nextVoucherIndexFromPage(prefix, year, page+1)
		if err != nil {
			return
		} else if nextPageIndex > index {
			index = nextPageIndex
		}
	}
	return
}

func (c *Client) hasNextPage(page string) bool {
	rxp, err := regexp.Compile("Datensatz \\d+ bis (\\d+) von (\\d+)")
	if err != nil {
		return false
	}
	n := rxp.FindStringSubmatch(page)
	if len(n) < 3 {
		return false
	}
	to, _ := strconv.Atoi(n[1])
	all, _ := strconv.Atoi(n[2])
	return to < all
}

func extractSubmatch(response *http.Response, regex *regexp.Regexp) (match string, page string, err error) {
	matches, page, err := extractSubmatches(response, regex)
	if err == nil {
		if len(matches) == 0 {
			err = errors.New("Not Found.")
		} else {
			match = matches[0]
		}
	}
	return
}

func extractSubmatches(response *http.Response, regex *regexp.Regexp) (matches []string, page string, err error) {
	var b bytes.Buffer
	if _, err = io.Copy(&b, response.Body); err != nil {
		return
	}
	page = b.String()
	// fmt.Println("Search in Text: ", page)
	fmt.Println("Searching for submatches: ", regex)
	submatches := regex.FindAllStringSubmatch(page, -1)
	matches = make([]string, 0)
	for _, sm := range submatches {
		if len(sm) >= 2 {
			matches = append(matches, sm[1])
		}
	}
	fmt.Println("Found matches: ", matches)
	return
}
