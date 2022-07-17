package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// BuidlCookieHeader buidl cookie header, for example, <name>=<value>[;expires=<date>][;domain=<domain_name>][;path=<path>][;secure][;httponly]
func BuildCookieHeader(name string, value string, expires *time.Time, domain string, path string, secure bool, httpOnly bool) string {
	data := []string{}
	data = append(data, fmt.Sprintf("%s=%s", name, value))
	if expires != nil {
		data = append(data, expires.UTC().String())
	}
	if domain != "" {
		data = append(data, fmt.Sprintf("domain=%s", domain))
	}
	if path != "" {
		data = append(data, fmt.Sprintf("path=%s", path))
	}
	if secure {
		data = append(data, "secure")
	}
	if httpOnly {
		data = append(data, "httponly")
	}
	return strings.Join(data, ";")
}

// ClientOptions http client options
type ClientOptions struct {
	Timeout int    `json:"timeout" yaml:"timeout"`
	CAFile  string `json:"caFile" yaml:"caFile"`
}

func NewClient(options ClientOptions, logger *logr.Logger) *Client {
	client := Client{options: options, logger: logger}
	return &client
}

type Client struct {
	options ClientOptions
	logger  *logr.Logger
}

func (c *Client) Get(target string, headers map[string]string, response interface{}) error {
	client := http.Client{Timeout: time.Duration(c.options.Timeout) * time.Second}
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		c.logger.Error(err, "failed to build http request")
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error(err, "failed to execute GET request", "status", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err, "failed to read GET response")
		return err
	}
	if resp.StatusCode >= 400 {
		var r HTTPError
		err = json.Unmarshal(body, &r)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return fmt.Errorf("[%d] %s", r.Code, r.Message)
	} else {
		err = json.Unmarshal(body, response)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return nil
	}
}

func (c *Client) Post(target string, headers map[string]string, request interface{}, response interface{}) error {
	client := http.Client{Timeout: time.Duration(c.options.Timeout) * time.Second}
	data, err := json.Marshal(request)
	if err != nil {
		c.logger.Error(err, "failed to marshal request")
		return err
	}
	req, err := http.NewRequest("POST", target, bytes.NewBuffer(data))
	if err != nil {
		c.logger.Error(err, "failed to build http request")
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error(err, "failed to execute POST request", "status", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err, "failed to read POST response")
		return err
	}
	if resp.StatusCode >= 400 {
		var r HTTPError
		err = json.Unmarshal(body, &r)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return fmt.Errorf("[%d] %s", r.Code, r.Message)
	} else if resp.StatusCode == 204 { // no content
		return nil
	} else {
		err = json.Unmarshal(body, response)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return nil
	}
}

func (c *Client) Delete(target string, headers map[string]string, response interface{}) error {
	client := http.Client{Timeout: time.Duration(c.options.Timeout) * time.Second}
	req, err := http.NewRequest("DELETE", target, nil)
	if err != nil {
		c.logger.Error(err, "failed to build http request")
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error(err, "failed to execute DELETE request", "status", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err, "failed to read DELETE response")
		return err
	}
	if resp.StatusCode >= 400 {
		var r HTTPError
		err = json.Unmarshal(body, &r)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return fmt.Errorf("[%d] %s", r.Code, r.Message)
	} else {
		err = json.Unmarshal(body, response)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return nil
	}
}

func (c *Client) Patch(target string, headers map[string]string, request interface{}, response interface{}) error {
	client := http.Client{Timeout: time.Duration(c.options.Timeout) * time.Second}
	data, err := json.Marshal(request)
	if err != nil {
		c.logger.Error(err, "failed to marshal request")
		return err
	}
	req, err := http.NewRequest("Patch", target, bytes.NewBuffer(data))
	if err != nil {
		c.logger.Error(err, "failed to build http request")
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error(err, "failed to execute Patch request", "status", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err, "failed to read Patch response")
		return err
	}
	if resp.StatusCode >= 400 {
		var r HTTPError
		err = json.Unmarshal(body, &r)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return fmt.Errorf("[%d] %s", r.Code, r.Message)
	} else if resp.StatusCode == 204 { // no content
		return nil
	} else {
		err = json.Unmarshal(body, response)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return nil
	}
}

func (c *Client) Put(target string, headers map[string]string, request interface{}, response interface{}) error {
	client := http.Client{Timeout: time.Duration(c.options.Timeout) * time.Second}
	data, err := json.Marshal(request)
	if err != nil {
		c.logger.Error(err, "failed to marshal request")
		return err
	}
	req, err := http.NewRequest("PUT", target, bytes.NewBuffer(data))
	if err != nil {
		c.logger.Error(err, "failed to build http request")
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error(err, "failed to execute PUT request", "status", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err, "failed to read PUT response")
		return err
	}
	if resp.StatusCode >= 400 {
		var r HTTPError
		err = json.Unmarshal(body, &r)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return fmt.Errorf("[%d] %s", r.Code, r.Message)
	} else if resp.StatusCode == 204 { // no content
		return nil
	} else {
		err = json.Unmarshal(body, response)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return nil
	}
}

func (c *Client) PostForm(target string, headers map[string]string, request map[string]string, response interface{}) error {
	data := make(map[string][]string)
	for k, v := range request {
		data[k] = []string{v}
	}

	resp, err := http.PostForm(target, data)

	if err != nil {
		c.logger.Error(err, "failed to execute POST Form request", "status", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err, "failed to read POST response")
		return err
	}
	if resp.StatusCode >= 400 {
		var r HTTPError
		err = json.Unmarshal(body, &r)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return fmt.Errorf("[%d] %s", r.Code, r.Message)
	} else if resp.StatusCode == 204 { // no content
		return nil
	} else {
		err = json.Unmarshal(body, response)
		if err != nil {
			c.logger.Error(err, "failed to unmarsshal response")
			return err
		}
		return nil
	}
}
