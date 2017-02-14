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

	"github.com/arangodb/testAgent/pkg/retry"
	"github.com/arangodb/testAgent/service/cluster"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
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

type ArangoResponse struct {
	StatusCode     int
	CoordinatorURL string
	Rev            string
}

func (e ArangoError) Error() string {
	return fmt.Sprintf("%s: (code %d, errorNum %d)", e.ErrorMessage, e.Code, e.ErrorNum)
}

func (c *ArangoClient) createURL(urlPath string, query url.Values) (string, url.URL, error) {
	if c.lastCoordinatorURL == nil {
		// Pick a random machine
		machines, err := c.cluster.Machines()
		if err != nil {
			return "", url.URL{}, maskAny(err)
		}
		if len(machines) == 0 {
			return "", url.URL{}, maskAny(fmt.Errorf("No machines available"))
		}
		index := rand.Intn(len(machines))
		u := machines[index].CoordinatorURL()
		c.lastCoordinatorURL = &u
	}
	lastCoordinatorURL := *c.lastCoordinatorURL
	u := lastCoordinatorURL
	u.Path = urlPath
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String(), lastCoordinatorURL, nil
}

// SetCoordinator sets the URL of the coordinator to use for the next request.
// Passing an empty string will use a random coordinator for the next request.
func (c *ArangoClient) SetCoordinator(coordinatorURL string) error {
	if coordinatorURL == "" {
		c.lastCoordinatorURL = nil
	} else {
		u, err := url.Parse(coordinatorURL)
		if err != nil {
			return maskAny(err)
		}
		c.lastCoordinatorURL = u
	}
	return nil
}

// Get performs a GET operation of a coordinator.
// If result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Get(urlPath string, query url.Values, header map[string]string, result interface{}, successStatusCodes, failureStatusCodes []int, operationTimeout, retryTimeout time.Duration) (ArangoResponse, error) {
	if aresp, err := c.requestWithRetry("GET", urlPath, query, header, nil, "", result, successStatusCodes, failureStatusCodes, operationTimeout, retryTimeout); err != nil {
		return aresp, maskAny(err)
	} else {
		return aresp, nil
	}
}

// Delete performs a DELETE operation of a coordinator.
func (c *ArangoClient) Delete(urlPath string, query url.Values, header map[string]string, successStatusCodes, failureStatusCodes []int, operationTimeout, retryTimeout time.Duration) (ArangoResponse, error) {
	if aresp, err := c.requestWithRetry("DELETE", urlPath, query, header, nil, "", nil, successStatusCodes, failureStatusCodes, operationTimeout, retryTimeout); err != nil {
		return aresp, maskAny(err)
	} else {
		return aresp, nil
	}
}

// Post performs a POST operation of a coordinator.
// The given input is posted to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Post(urlPath string, query url.Values, header map[string]string, input interface{}, contentType string, result interface{}, successStatusCodes, failureStatusCodes []int, operationTimeout, retryTimeout time.Duration) (ArangoResponse, error) {
	if aresp, err := c.requestWithRetry("POST", urlPath, query, header, input, contentType, result, successStatusCodes, failureStatusCodes, operationTimeout, retryTimeout); err != nil {
		return aresp, maskAny(err)
	} else {
		return aresp, nil
	}
}

// Patch performs a PATCH operation on a coordinator.
// The given input is send to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Patch(urlPath string, query url.Values, header map[string]string, input interface{}, contentType string, result interface{}, successStatusCodes, failureStatusCodes []int, operationTimeout, retryTimeout time.Duration) (ArangoResponse, error) {
	if aresp, err := c.requestWithRetry("PATCH", urlPath, query, header, input, contentType, result, successStatusCodes, failureStatusCodes, operationTimeout, retryTimeout); err != nil {
		return aresp, maskAny(err)
	} else {
		return aresp, nil
	}
}

// Put performs a PUT operation on a coordinator.
// The given input is send to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Put(urlPath string, query url.Values, header map[string]string, input interface{}, contentType string, result interface{}, successStatusCodes, failureStatusCodes []int, operationTimeout, retryTimeout time.Duration) (ArangoResponse, error) {
	if aresp, err := c.requestWithRetry("PUT", urlPath, query, header, input, contentType, result, successStatusCodes, failureStatusCodes, operationTimeout, retryTimeout); err != nil {
		return aresp, maskAny(err)
	} else {
		return aresp, nil
	}
}

// requestWithRetry performs an operation on a coordinator.
// The given input is send to the server (if any), if result != nil and status is success, the response is parsed into result.
func (c *ArangoClient) requestWithRetry(method, urlPath string, query url.Values, header map[string]string, input interface{}, contentType string, result interface{}, successStatusCodes, failureStatusCodes []int, operationTimeout, retryTimeout time.Duration) (ArangoResponse, error) {
	inputData, contentType, err := prepareInput(input, contentType)
	if err != nil {
		return ArangoResponse{}, maskAny(err)
	}
	attempt := 0
	var aresp ArangoResponse
	op := func() error {
		attempt++
		start := time.Now()
		client := createClient(operationTimeout)
		url, lastCoordinatorURL, err := c.createURL(urlPath, query)
		if err != nil {
			return maskAny(errors.Wrapf(err, "Failed creating URL for path '%s' (attempt %d, after %s, error %v)", urlPath, attempt, time.Since(start), err))
		}
		aresp.CoordinatorURL = lastCoordinatorURL.String()
		var rd io.Reader
		if inputData != nil {
			rd = bytes.NewReader(inputData)
		}
		req, err := http.NewRequest(method, url, rd)
		if err != nil {
			return maskAny(errors.Wrapf(err, "Failed creating %s request for path '%s' (attempt %d, after %s, error %v)", method, urlPath, attempt, time.Since(start), err))
		}
		if inputData != nil {
			req.Header.Set("Content-Type", contentType)
		}
		for k, v := range header {
			req.Header.Set(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			c.lastCoordinatorURL = nil // Change coordinator
			return maskAny(errors.Wrapf(err, "Failed performing %s request to %s (attempt %d, after %s, error %v)", method, url, attempt, time.Since(start), err))
		}
		// Process response
		if err := c.handleResponse(resp, method, url, result, &aresp, successStatusCodes, failureStatusCodes, attempt, start); err != nil {
			return maskAny(err)
		}
		return nil
	}

	if err := retry.Retry(op, retryTimeout); err != nil {
		return aresp, maskAny(err)
	}
	return aresp, nil
}

func (c *ArangoClient) handleResponse(resp *http.Response, method, url string, result interface{}, aresp *ArangoResponse, successStatusCodes, failureStatusCodes []int, attempt int, start time.Time) error {
	// Store status code
	aresp.StatusCode = resp.StatusCode

	// Read response body into memory
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return maskAny(errors.Wrapf(err, "Failed reading response data from %s request to %s (attempt %d, after %s, error %v)", method, url, attempt, time.Since(start), err))
	}

	// Check for failure status
	for _, code := range failureStatusCodes {
		if resp.StatusCode == code {
			var aerr ArangoError
			headers := formatHeaders(resp)
			if tryDecodeBody(body, &aerr); err == nil {
				return maskAny(errors.Wrapf(retry.FailureError, "Received status %d, from %s request to %s, which is a failure (attempt %d, after %s, error %s, headers\n%s\n)", resp.StatusCode, method, url, attempt, time.Since(start), aerr.Error(), headers))
			}
			return maskAny(errors.Wrapf(retry.FailureError, "Received status %d, from %s request to %s, which is a failure (attempt %d, after %s, headers\n%s\n\nbody\n%s\n)", resp.StatusCode, method, url, attempt, time.Since(start), headers, string(body)))
		}
	}

	// Check for success status
	for _, code := range successStatusCodes {
		if resp.StatusCode == code {
			// Fetch update headers
			aresp.Rev = resp.Header.Get("etag")

			// Found a success status
			if isSuccessStatusCode(code) && result != nil {
				if err := json.Unmarshal(body, result); err != nil {
					return maskAny(errors.Wrapf(err, "Failed decoding response data from %s request to %s (attempt %d, after %s, error %v)", method, url, attempt, time.Since(start), err))
				}
			}
			// Return success
			return nil
		}
	}

	// Unexpected status code
	c.lastCoordinatorURL = nil // Change coordinator
	headers := formatHeaders(resp)
	return maskAny(fmt.Errorf("Unexpected status %d from %s request to %s (attempt %d, after %s, headers\n%s\n\nbody\n%s\n)", resp.StatusCode, method, url, attempt, time.Since(start), headers, string(body)))
}

func tryDecodeBody(body []byte, result interface{}) error {
	if err := json.Unmarshal(body, result); err != nil {
		return maskAny(err)
	}
	return nil
}

func prepareInput(input interface{}, contentType string) ([]byte, string, error) {
	if contentType == "" {
		contentType = contentTypeJson
	}
	if input == nil {
		return nil, contentType, nil
	}
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
