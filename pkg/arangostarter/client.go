package arangostarter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/arangodb/testAgent/pkg/retry"
	"github.com/pkg/errors"
)

func NewArangoStarterClient(ip string, port int) (API, error) {
	endpoint, err := url.Parse(fmt.Sprintf("http://%s:%d", ip, port))
	if err != nil {
		return nil, maskAny(err)
	}
	return &client{
		endpoint: *endpoint,
		client:   &http.Client{Timeout: time.Second * 15},
	}, nil
}

type client struct {
	endpoint url.URL
	client   *http.Client
}

type HelloRequest struct {
	SlaveID      string // Unique ID of the slave
	SlaveAddress string // Address used to reach the slave (if empty, this will be derived from the request)
	SlavePort    int    // Port used to reach the slave
	DataDir      string // Directory used for data by this slave
}

const (
	contentTypeJSON = "application/json"
)

func (c *client) PrepareSlave(id, hostIP string, port int, dataDir string) (Peer, error) {
	helloReq := HelloRequest{
		SlaveID:      id,
		SlaveAddress: hostIP,
		SlavePort:    port,
		DataDir:      dataDir,
	}
	input, err := json.Marshal(helloReq)
	if err != nil {
		return Peer{}, maskAny(err)
	}
	url := c.createURL("/hello", nil)

	var result Peer
	op := func() error {
		resp, err := c.client.Post(url, contentTypeJSON, bytes.NewReader(input))
		if err != nil {
			return maskAny(err)
		}
		var peers Peers
		if err := c.handleResponse(resp, "POST", url, &peers); err != nil {
			return maskAny(err)
		}

		p, ok := peers.PeerByID(id)
		if !ok {
			return maskAny(fmt.Errorf("Master did not include peer '%s'", id))
		}
		result = p
		return nil
	}

	if err := retry.Retry(op, time.Minute*2); err != nil {
		return Peer{}, maskAny(err)
	}

	return result, nil
}

// updateServerInfo connects to arangodb to query the port numbers & container info
// of all servers on the machine
func (c *client) GetProcesses() (ProcessListResponse, error) {
	url := c.createURL("/process", nil)

	var result ProcessListResponse
	op := func() error {
		resp, err := c.client.Get(url)
		if err != nil {
			return maskAny(err)
		}
		if err := c.handleResponse(resp, "GET", url, &result); err != nil {
			return maskAny(err)
		}
		return nil
	}

	if err := retry.Retry(op, time.Minute*2); err != nil {
		return ProcessListResponse{}, maskAny(err)
	}

	return result, nil
}

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

func (c *client) createURL(urlPath string, query url.Values) string {
	u := c.endpoint
	u.Path = urlPath
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String()
}
