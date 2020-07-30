package http

import (
	"bytes"
	"fmt"
	"github.com/marsmay/golib/strings2"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Client struct {
	client *http.Client
}

func (c *Client) Do(url string, header map[string]string, body []byte) ([]byte, error) {
	method := strings2.IIf(body == nil, "GET", "POST")
	req, err := http.NewRequest(method, url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error http code %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *Client) DoMethod(url string, header map[string]string, body []byte, method string) ([]byte, error) {
	method = strings2.IIf(method == "", "GET", method)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error http code %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func NewClient(timeout time.Duration) *Client {
	cookie, _ := cookiejar.New(nil)
	return &Client{client: &http.Client{Jar: cookie, Timeout: timeout}}
}
