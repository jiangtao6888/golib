package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Client struct {
	client *http.Client
}

func (c *Client) Get(url string, header map[string]string, body []byte) (int, []byte, error) {
	return c.Do(http.MethodGet, url, header, body, true)
}

func (c *Client) Post(url string, header map[string]string, body []byte) (int, []byte, error) {
	return c.Do(http.MethodPost, url, header, body, true)
}

func (c *Client) Do(method string, url string, header map[string]string, body []byte, checkStatus bool) (respCode int, respBody []byte, err error) {
	var req *http.Request

	if len(body) > 0 {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respCode = resp.StatusCode

	if checkStatus && respCode != 200 {
		err = fmt.Errorf("error http code %d", respCode)
		return
	}

	respBody, err = ioutil.ReadAll(resp.Body)
	return
}

func NewClient(timeout time.Duration) *Client {
	cookie, _ := cookiejar.New(nil)
	return &Client{client: &http.Client{Jar: cookie, Timeout: timeout}}
}
