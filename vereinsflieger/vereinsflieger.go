package vereinsflieger

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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

func (c *Client) PostFormString(url string, vals url.Values) (body string, err error) {
	resp, err := c.PostForm(url, vals)
	if err != nil {
		return
	}
	body, err = readIntoString(resp.Body)
	return
}

func (c *Client) GetString(url string) (body string, err error) {
	resp, err := c.Get(url)
	if err != nil {
		return
	}
	body, err = readIntoString(resp.Body)
	return
}

func readIntoString(r io.Reader) (str string, err error) {
	b := bytes.Buffer{}
	_, err = io.Copy(&b, r)
	if err != nil {
		return
	}
	str = b.String()
	return
}
