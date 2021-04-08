package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	//"github.com/arangodb-helper/testagent/pkg/retry"
	"github.com/arangodb-helper/testagent/service/cluster"
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
	startTSFormat   = "2006-01-02 15:04:05"
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
	StatusCode     int    // HTTP status code of last attempt
	CoordinatorURL string // URL of coordinator used for last attempt
	Rev            string // Revision of document as returned by database (not set for all operations)
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
func (c *ArangoClient) Get(
	urlPath string, query url.Values, header map[string]string, result interface{}, successStatusCodes,
	failureStatusCodes []int, operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	return c.requestWithRetry("GET", urlPath, query, header, nil, "", result,
		successStatusCodes, failureStatusCodes, operationTimeout, retries)
}

// Delete performs a DELETE operation of a coordinator.
func (c *ArangoClient) Delete(
	urlPath string, query url.Values, header map[string]string, successStatusCodes,
	failureStatusCodes []int, operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	return c.requestWithRetry("DELETE", urlPath, query, header, nil, "", nil,
		successStatusCodes, failureStatusCodes, operationTimeout, retries)
}

// Post performs a POST operation of a coordinator.
// The given input is posted to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Post(
	urlPath string, query url.Values, header map[string]string, input interface{},
	contentType string, result interface{}, successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	return c.requestWithRetry("POST", urlPath, query, header, input, contentType,
		result, successStatusCodes, failureStatusCodes, operationTimeout, retries)
}

// Patch performs a PATCH operation on a coordinator.
// The given input is send to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Patch(
	urlPath string, query url.Values, header map[string]string, input interface{},
	contentType string, result interface{}, successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	return c.requestWithRetry("PATCH", urlPath, query, header, input, contentType,
		result, successStatusCodes, failureStatusCodes, operationTimeout, retries)
}

// Put performs a PUT operation on a coordinator.
// The given input is send to the server, if result != nil and status == 200, the response is parsed into result.
func (c *ArangoClient) Put(
	urlPath string, query url.Values, header map[string]string, input interface{},
	contentType string, result interface{}, successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	return c.requestWithRetry("PUT", urlPath, query, header, input, contentType,
		result, successStatusCodes, failureStatusCodes, operationTimeout, retries)
}

// requestWithRetry performs an operation on a coordinator.
// The given input is send to the server (if any), if result != nil and status is success, the response is parsed into result.
func (c *ArangoClient) requestWithRetry(
	method, urlPath string, query url.Values, header map[string]string, input interface{},
	contentType string, result interface{}, successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {

	aresps := make([]ArangoResponse, 0, retries)
	errors := make([]error, 0, retries)

	inputData, contentType, err := prepareInput(input, contentType)
	if err != nil {
		aresps = append(aresps, ArangoResponse{})
		errors = append(errors, maskAny(err))
		return aresps, errors
	}

	var i int

	op := func() (ArangoResponse, error) {
		var arangoResp ArangoResponse
		start := time.Now()
		client := createClient(operationTimeout)
		url, lastCoordinatorURL, err := c.createURL(urlPath, query)
		if err != nil {
			return arangoResp, fmt.Errorf("Failed creating URL for path '%s' (attempt %d, started at %s, after %s, error %v)", urlPath, i, start.Format(startTSFormat), time.Since(start), err)
		}
		arangoResp.CoordinatorURL = lastCoordinatorURL.String()
		var rd io.Reader
		if inputData != nil {
			rd = bytes.NewReader(inputData)
		}
		req, err := http.NewRequest(method, url, rd)
		if err != nil {
			return arangoResp, fmt.Errorf("Failed creating %s request for path '%s' (attempt %d, started at %s, after %s, error %v)", method, urlPath, i, start.Format(startTSFormat), time.Since(start), err)
		}
		if inputData != nil {
			req.Header.Set("Content-Type", contentType)
		}
		for k, v := range header {
			req.Header.Set(k, v)
		}
		httpResp, err := client.Do(req)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				arangoResp.StatusCode = 0
			} else {
				if strings.Contains(err.Error(), "refused") {
					arangoResp.StatusCode = 1
				} else if strings.Contains(err.Error(), "canceled") ||
					strings.Contains(err.Error(), "context deadline exceeded") {
					arangoResp.StatusCode = 0
				}
			}
			c.lastCoordinatorURL = nil // Change coordinator
			return arangoResp, nil
		}
		// Process response
		if err := c.handleResponse(httpResp, method, url, result, &arangoResp, i, successStatusCodes, failureStatusCodes, start); err != nil {
			return arangoResp, maskAny(err)
		}
		return arangoResp, nil
	}

	for i = 0; i < retries; i++ {
		aresp, err := op()
		aresps = append(aresps, aresp)
		errors = append(errors, maskAny(err))
		if err == nil {
			break
		}
	}

	return aresps, errors
}

func (c *ArangoClient) handleResponse(
	resp *http.Response, method, url string, result interface{}, aresp *ArangoResponse,
	attempt int, successStatusCodes, failureStatusCodes []int, start time.Time) error {
	// Store status code
	aresp.StatusCode = resp.StatusCode

	// Read response body into memory
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return maskAny(errors.Wrapf(err, "Failed reading response data from %s request to %s (attempt %d, started at %s, after %s, error %v)", method, url, attempt, start.Format(startTSFormat), time.Since(start), err))
	}

	// Check for failure status
	for _, code := range failureStatusCodes {
		if resp.StatusCode == code {
			var aerr ArangoError
			headers := formatHeaders(resp)
			if err := tryDecodeBody(body, &aerr); err == nil {
				return maskAny(errors.Wrapf(err, "Received status %d, from %s request to %s, which is a failure (attempt %d, started at %s, after %s, error %s, headers\n%s\n)", resp.StatusCode, method, url, attempt, start.Format(startTSFormat), time.Since(start), aerr.Error(), headers))
			}
			return maskAny(errors.Wrapf(err, "Received status %d, from %s request to %s, which is a failure (attempt %d, started at %s, after %s, headers\n%s\n\nbody\n%s\n)", resp.StatusCode, method, url, attempt, start.Format(startTSFormat), time.Since(start), headers, string(body)))
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
					return maskAny(errors.Wrapf(err, "Failed decoding response data from %s request to %s (attempt %d, started at %s, after %s, error %v)", method, url, attempt, start.Format(startTSFormat), time.Since(start), err))
				}
			}
			// Return success
			return nil
		}
	}

	// Unexpected status code
	c.lastCoordinatorURL = nil // Change coordinator
	headers := formatHeaders(resp)
	return maskAny(fmt.Errorf("Unexpected status %d from %s request to %s (attempt %d, started at %s, after %s, headers\n%s\n\nbody\n%s\n)", resp.StatusCode, method, url, attempt, start.Format(startTSFormat), time.Since(start), headers, string(body)))
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
