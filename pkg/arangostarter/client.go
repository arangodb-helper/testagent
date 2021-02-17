//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package arangostarter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/arangodb-helper/testagent/pkg/retry"
)

// NewArangoStarterClient creates a new client implementation.
func NewArangoStarterClient(endpoint url.URL) (API, error) {
	endpoint.Path = ""
	return &client{
		endpoint: endpoint,
		client:   shardHTTPClient,
	}, nil
}

var (
	shardHTTPClient = DefaultHTTPClient()
)

type client struct {
	endpoint url.URL
	client   *http.Client
}

const (
	contentTypeJSON = "application/json"
)

// ID requests the starters ID.
func (c *client) ID(ctx context.Context) (IDInfo, error) {
	url := c.createURL("/id", nil)

	var result IDInfo
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return IDInfo{}, maskAny(err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return IDInfo{}, maskAny(err)
	}
	if err := c.handleResponse(resp, "GET", url, &result); err != nil {
		return IDInfo{}, maskAny(err)
	}

	return result, nil
}

// Version requests the starter version.
func (c *client) Version(ctx context.Context) (VersionInfo, error) {
	url := c.createURL("/version", nil)

	var result VersionInfo
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return VersionInfo{}, maskAny(err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return VersionInfo{}, maskAny(err)
	}
	if err := c.handleResponse(resp, "GET", url, &result); err != nil {
		return VersionInfo{}, maskAny(err)
	}

	return result, nil
}

// Processes loads information of all the server processes launched by a specific arangodb.
func (c *client) Processes(ctx context.Context) (ProcessList, error) {
	url := c.createURL("/process", nil)

	var result ProcessList
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ProcessList{}, maskAny(err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return ProcessList{}, maskAny(err)
	}
	if err := c.handleResponse(resp, "GET", url, &result); err != nil {
		return ProcessList{}, maskAny(err)
	}

	return result, nil
}

// Endpoints loads the URL's needed to reach all starters, agents & coordinators in the cluster.
func (c *client) Endpoints(ctx context.Context) (EndpointList, error) {
	url := c.createURL("/endpoints", nil)

	var result EndpointList
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return EndpointList{}, maskAny(err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return EndpointList{}, maskAny(err)
	}
	if err := c.handleResponse(resp, "GET", url, &result); err != nil {
		return EndpointList{}, maskAny(err)
	}

	return result, nil
}

// Shutdown will shutdown a starter (and all its started servers).
// With goodbye set, it will remove the peer slot for the starter.
func (c *client) Shutdown(ctx context.Context, goodbye bool) error {
	q := url.Values{}
	if goodbye {
		q.Set("mode", "goodbye")
	}
	url := c.createURL("/shutdown", q)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return maskAny(err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return maskAny(err)
	}
	if err := c.handleResponse(resp, "POST", url, nil); err != nil {
		return maskAny(err)
	}

	return nil
}

// WaitUntilGone will block until the specific arangodb instance can no longer be reached.
func (c *client) WaitUntilGone(ctx context.Context) error {
	url := c.createURL("/version", nil)

	var result VersionInfo
	op := func() error {
		resp, err := c.client.Get(url)
		if err != nil {
			// When the arangodb is gone, we expect an error here.
			// In that case we're done
			return nil
		}
		if err := c.handleResponse(resp, "POST", url, &result); err != nil {
			// When the arangodb is gone, we expect an error here.
			// In that case we're done
			return nil
		}
		return maskAny(fmt.Errorf("Still not gone"))
	}

	if err := retry.Retry(op, time.Minute*10); err != nil {
		return maskAny(err)
	}

	return nil
}

// handleResponse checks the given response status and decodes any JSON result.
func (c *client) handleResponse(resp *http.Response, method, url string, result interface{}) error {
	// Read response body into memory
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return maskAny(errors.Wrapf(err, "Failed reading response data from %s request to %s: %v", method, url, err))
	}

	if resp.StatusCode != http.StatusOK {
		/*var er ErrorResponse
		if err := json.Unmarshal(body, &er); err == nil {
			return &er
		}*/
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

// createURL creates a full URL for a request with given local path & query.
func (c *client) createURL(urlPath string, query url.Values) string {
	u := c.endpoint
	u.Path = urlPath
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String()
}
