package networkblocker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type client struct {
	endpoint url.URL
	client   *http.Client
}

func NewClient(endpoint url.URL) API {
	return &client{
		endpoint: endpoint,
		client:   &http.Client{Timeout: time.Second * 15},
	}
}

const (
	contentTypeJSON = "application/json"
)

type ErrorResponse struct {
	Message string `json:"error,omitempty"`
}

type RulesResponse struct {
	Rules []string `json:"rules,omitempty"`
}

func (er *ErrorResponse) Error() string {
	return er.Message
}

// RejectTCP actively denies all traffic on the given TCP port
func (c *client) RejectTCP(port int) error {
	url := c.createURL(fmt.Sprintf("/api/v1/reject/tcp/%d", port), nil)
	resp, err := c.client.Post(url, contentTypeJSON, nil)
	if err != nil {
		return maskAny(err)
	}
	if err := c.handleResponse(resp, "POST", url, nil); err != nil {
		return maskAny(err)
	}
	return nil
}

// DropTCP silently denies all traffic on the given TCP port
func (c *client) DropTCP(port int) error {
	url := c.createURL(fmt.Sprintf("/api/v1/drop/tcp/%d", port), nil)
	resp, err := c.client.Post(url, contentTypeJSON, nil)
	if err != nil {
		return maskAny(err)
	}
	if err := c.handleResponse(resp, "POST", url, nil); err != nil {
		return maskAny(err)
	}
	return nil
}

// AcceptTCP allow all traffic on the given TCP port
func (c *client) AcceptTCP(port int) error {
	url := c.createURL(fmt.Sprintf("/api/v1/accept/tcp/%d", port), nil)
	resp, err := c.client.Post(url, contentTypeJSON, nil)
	if err != nil {
		return maskAny(err)
	}
	if err := c.handleResponse(resp, "POST", url, nil); err != nil {
		return maskAny(err)
	}
	return nil
}

// Rules returns a list of all rules injected by this service.
func (c *client) Rules() ([]string, error) {
	url := c.createURL("/api/v1/rules", nil)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, maskAny(err)
	}
	var data RulesResponse
	if err := c.handleResponse(resp, "GET", url, &data); err != nil {
		return nil, maskAny(err)
	}
	return data.Rules, nil

}

func (c *client) handleResponse(resp *http.Response, method, url string, result interface{}) error {
	// Read response body into memory
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return maskAny(errors.Wrapf(err, "Failed reading response data from %s request to %s: %v", method, url, err))
	}

	if resp.StatusCode != http.StatusOK {
		var er ErrorResponse
		if err := json.Unmarshal(body, &er); err == nil {
			return &er
		}
		return maskAny(fmt.Errorf("Invalid status %d", resp.StatusCode))
	}

	// Got a success status
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return maskAny(errors.Wrapf(err, "Failed decoding response data from %s request to %s: %v", method, url, err))
		}
	}
	return nil
}

func (c *client) createURL(urlPath string, query url.Values) string {
	u := c.endpoint
	u.Path = urlPath
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String()
}
