package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/arangodb/testAgent/service/cluster"
	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
)

func NewArangoClient(log *logging.Logger, cluster cluster.Cluster) *ArangoClient {
	return &ArangoClient{
		log:     log,
		cluster: cluster,
	}
}

const (
	contentTypeJson = "application/json"
)

type ArangoClient struct {
	log                *logging.Logger
	cluster            cluster.Cluster
	lastCoordinatorURL *url.URL
}

type ArangoError struct {
	Error_       bool   `json:"error,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Code         int    `json:"code,omitempty"`
	ErrorNum     int    `json:"errorNum,omitempty"`
}

func (e ArangoError) Error() string {
	return fmt.Sprintf("%s: (code %d, errorNum %d)", e.ErrorMessage, e.Code, e.ErrorNum)
}

func (c *ArangoClient) createURL(urlPath string, query url.Values) (string, error) {
	if c.lastCoordinatorURL == nil {
		// Pick a random machine
		machines, err := c.cluster.Machines()
		if err != nil {
			return "", maskAny(err)
		}
		if len(machines) == 0 {
			return "", maskAny(fmt.Errorf("No machines available"))
		}
		index := rand.Intn(len(machines))
		u := machines[index].CoordinatorURL()
		c.lastCoordinatorURL = &u
	}
	u := *c.lastCoordinatorURL
	u.Path = urlPath
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

// Get performs a GET operation of a coordinator.
// If result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Get(urlPath string, query url.Values, result interface{}, successStatusCodes, failureStatusCodes []int, timeout time.Duration) error {
	op := func() error {
		client := createClient(timeout)
		url, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating URL for path '%s': %v", urlPath, err))
		}
		resp, err := client.Get(url)
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errgo.WithCausef(nil, err, "Failed performing GET request to %s: %v", url, err))
		}
		// Process response
		if err := c.handleResponse(resp, "GET", url, result, successStatusCodes, failureStatusCodes); err != nil {
			return maskAny(err)
		}
		return nil
	}

	if err := retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Delete performs a DELETE operation of a coordinator.
func (c *ArangoClient) Delete(urlPath string, query url.Values, successStatusCodes, failureStatusCodes []int, timeout time.Duration) error {
	op := func() error {
		client := createClient(timeout)
		url, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating URL for path '%s': %v", urlPath, err))
		}
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating DELETE request for path '%s': %v", urlPath, err))
		}
		resp, err := client.Do(req)
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errgo.WithCausef(nil, err, "Failed performing DELETE request to %s: %v", url, err))
		}
		// Process response
		if err := c.handleResponse(resp, "DELETE", url, nil, successStatusCodes, failureStatusCodes); err != nil {
			return maskAny(err)
		}
		return nil
	}

	if err := retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Post performs a POST operation of a coordinator.
// The given input is posted to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Post(urlPath string, query url.Values, input interface{}, contentType string, result interface{}, successStatusCodes, failureStatusCodes []int, timeout time.Duration) error {
	inputData, contentType, err := prepareInput(input, contentType)
	if err != nil {
		return maskAny(err)
	}
	op := func() error {
		client := createClient(timeout)
		url, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating URL for path '%s': %v", urlPath, err))
		}
		resp, err := client.Post(url, contentType, bytes.NewReader(inputData))
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errgo.WithCausef(nil, err, "Failed performing POST request to %s: %v", url, err))
		}
		// Process response
		if err := c.handleResponse(resp, "POST", url, result, successStatusCodes, failureStatusCodes); err != nil {
			return maskAny(err)
		}
		return nil
	}

	if err := retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Patch performs a PATCH operation on a coordinator.
// The given input is send to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Patch(urlPath string, query url.Values, input interface{}, contentType string, result interface{}, successStatusCodes, failureStatusCodes []int, timeout time.Duration) error {
	inputData, contentType, err := prepareInput(input, contentType)
	if err != nil {
		return maskAny(err)
	}
	op := func() error {
		client := createClient(timeout)
		url, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating URL for path '%s': %v", urlPath, err))
		}
		req, err := http.NewRequest("PATCH", url, bytes.NewReader(inputData))
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating DELETE request for path '%s': %v", urlPath, err))
		}
		req.Header.Set("Content-Type", contentType)
		resp, err := client.Do(req)
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errgo.WithCausef(nil, err, "Failed performing PATCH request to %s: %v", url, err))
		}
		// Process response
		if err := c.handleResponse(resp, "PATCH", url, result, successStatusCodes, failureStatusCodes); err != nil {
			return maskAny(err)
		}
		return nil
	}

	if err := retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

func (c *ArangoClient) handleResponse(resp *http.Response, method, url string, result interface{}, successStatusCodes, failureStatusCodes []int) error {
	// Read response body into memory
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return maskAny(errgo.WithCausef(nil, err, "Failed reading response data from %s request to %s: %v", method, url, err))
	}

	// Check for failure status
	for _, code := range failureStatusCodes {
		if resp.StatusCode == code {
			var aerr ArangoError
			headers := formatHeaders(resp)
			if tryDecodeBody(body, &aerr); err == nil {
				return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from %s request to %s, which is a failure (%s); headers:\n%s", resp.StatusCode, method, url, aerr.Error(), headers))
			}
			return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from %s request to %s, which is a failure; headers:\n%s\n\nbody:\n%s", resp.StatusCode, method, url, headers, string(body)))
		}
	}

	// Check for success status
	for _, code := range successStatusCodes {
		if resp.StatusCode == code {
			// Found a success status
			if isSuccessStatusCode(code) && result != nil {
				if err := json.Unmarshal(body, result); err != nil {
					return maskAny(errgo.WithCausef(nil, err, "Failed decoding response data from %s request to %s: %v", method, url, err))
				}
			}
			// Return success
			return nil
		}
	}

	// Unexpected status code
	c.lastCoordinatorURL = nil // Change coordinator
	headers := formatHeaders(resp)
	return maskAny(fmt.Errorf("Unexpected status %d from %s request to %s; headers:\n%s\n\nbody:\n%s", resp.StatusCode, method, url, headers, string(body)))
}

func tryDecodeBody(body []byte, result interface{}) error {
	if err := json.Unmarshal(body, result); err != nil {
		return maskAny(err)
	}
	return nil
}

func prepareInput(input interface{}, contentType string) ([]byte, string, error) {
	inputData, ok := input.([]byte)
	if !ok {
		var err error
		if rd, ok := input.(io.Reader); ok {
			inputData, err = ioutil.ReadAll(rd)
			if err != nil {
				return nil, "", maskAny(err)
			}
		} else {
			inputData, err = json.Marshal(input)
			if err != nil {
				return nil, "", maskAny(err)
			}
		}
	}
	if contentType == "" {
		contentType = contentTypeJson
	}
	return inputData, contentType, nil
}

func formatHeaders(resp *http.Response) string {
	buf := &bytes.Buffer{}
	resp.Header.Write(buf)
	return buf.String()
}

func isSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func createClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}
