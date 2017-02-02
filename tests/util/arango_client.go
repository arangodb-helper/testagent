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
		url, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating URL for path '%s': %v", urlPath, err))
		}
		resp, err := http.Get(url)
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errgo.WithCausef(nil, err, "Failed performing GET request to %s: %v", url, err))
		}
		// Check for failure status
		for _, code := range failureStatusCodes {
			if resp.StatusCode == code {
				var aerr ArangoError
				if tryDecodeBody(resp.Body, &aerr); err == nil {
					return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from GET request to %s, which is a failure (%s)", resp.StatusCode, url, aerr.Error()))
				}
				return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from GET request to %s, which is a failure", resp.StatusCode, url))
			}
		}
		// Check for success status
		for _, code := range successStatusCodes {
			if resp.StatusCode == code {
				// Found a success status
				if isSuccessStatusCode(code) && result != nil {
					defer resp.Body.Close()
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return maskAny(errgo.WithCausef(nil, err, "Failed reading response data from GET request to %s: %v", url, err))
					}
					if err := json.Unmarshal(b, result); err != nil {
						return maskAny(errgo.WithCausef(nil, err, "Failed decoding response data from GET request to %s: %v", url, err))
					}
				}
				// Return success
				return nil
			}
		}
		// Unexpected status code
		c.lastCoordinatorURL = nil // Change coordinator
		return maskAny(fmt.Errorf("Unexpected status %d from GET request to %s", resp.StatusCode, url))
	}

	if err := retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Delete performs a DELETE operation of a coordinator.
func (c *ArangoClient) Delete(urlPath string, query url.Values, successStatusCodes, failureStatusCodes []int, timeout time.Duration) error {
	op := func() error {
		url, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating URL for path '%s': %v", urlPath, err))
		}
		resp, err := http.Get(url)
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errgo.WithCausef(nil, err, "Failed performing DELETE request to %s: %v", url, err))
		}
		// Check for failure status
		for _, code := range failureStatusCodes {
			if resp.StatusCode == code {
				var aerr ArangoError
				if tryDecodeBody(resp.Body, &aerr); err == nil {
					return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from DELETE request to %s, which is a failure (%s)", resp.StatusCode, url, aerr.Error()))
				}
				return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from DELETE request to %s, which is a failure", resp.StatusCode, url))
			}
		}
		// Check for success status
		for _, code := range successStatusCodes {
			if resp.StatusCode == code {
				// Return success
				return nil
			}
		}
		// Unexpected status code
		c.lastCoordinatorURL = nil // Change coordinator
		return maskAny(fmt.Errorf("Unexpected status %d from DELETE request to %s", resp.StatusCode, url))
	}

	if err := retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Post performs a POST operation of a coordinator.
// The given input is posted to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Post(urlPath string, query url.Values, input interface{}, contentType string, result interface{}, successStatusCodes, failureStatusCodes []int, timeout time.Duration) error {
	inputReader, ok := input.(io.Reader)
	if !ok {
		body, err := json.Marshal(input)
		if err != nil {
			return maskAny(err)
		}
		inputReader = bytes.NewReader(body)
	}
	if contentType == "" {
		contentType = contentTypeJson
	}
	op := func() error {
		url, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed creating URL for path '%s': %v", urlPath, err))
		}
		resp, err := http.Post(url, contentType, inputReader)
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errgo.WithCausef(nil, err, "Failed performing POST request to %s: %v", url, err))
		}
		// Check for failure status
		for _, code := range failureStatusCodes {
			if resp.StatusCode == code {
				var aerr ArangoError
				if tryDecodeBody(resp.Body, &aerr); err == nil {
					return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from POST request to %s, which is a failure (%s)", resp.StatusCode, url, aerr.Error()))
				}
				return maskAny(errgo.WithCausef(nil, failureError, "Received status %d, from POST request to %s, which is a failure", resp.StatusCode, url))
			}
		}
		// Check for success status
		for _, code := range successStatusCodes {
			if resp.StatusCode == code {
				// Found a success status
				if isSuccessStatusCode(code) && result != nil {
					defer resp.Body.Close()
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return maskAny(errgo.WithCausef(nil, err, "Failed reading response data from POST request to %s: %v", url, err))
					}
					if err := json.Unmarshal(b, result); err != nil {
						return maskAny(errgo.WithCausef(nil, err, "Failed decoding response data from POST request to %s: %v", url, err))
					}
				}
				// Return success
				return nil
			}
		}
		// Unexpected status code
		c.lastCoordinatorURL = nil // Change coordinator
		return maskAny(fmt.Errorf("Unexpected status %d from POST request to %s", resp.StatusCode, url))
	}

	if err := retry(op, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}

func tryDecodeBody(body io.ReadCloser, result interface{}) error {
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return maskAny(err)
	}
	if err := json.Unmarshal(b, result); err != nil {
		return maskAny(err)
	}
	return nil
}

func isSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
