package docker

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"

)

type DockerHost struct {
	Client    *dc.Client
	IP        string
	Endpoint  string
	Interface string
}

var (
	maskAny = errors.WithStack
)

// NewDockerHosts creates DockerHosts for each of the given endpoints.
// When an endpoint is a unix socket, the given localHostIP is assumed to be its host IP.
func NewDockerHosts(endpoints []string, localHostIP, dockerIntf string) ([]*DockerHost, error) {
	var dockerHosts []*DockerHost
	for _, endpoint := range endpoints {
		hostIP, err := getHostAddressForEndpoint(endpoint, localHostIP)
		if err != nil {
			return nil, maskAny(err)
		}

		// Create docker host
		dockerHost, err := newDockerHost(endpoint, hostIP, dockerIntf)
		if err != nil {
			return nil, maskAny(err)
		}
		dockerHosts = append(dockerHosts, dockerHost)
		time.Sleep(5 * time.Second)
	}
	return dockerHosts, nil
}

func newDockerHost(endpoint, hostIP, intf string) (*DockerHost, error) {

	os.Setenv("DOCKER_HOST", endpoint)
	client, err := dc.NewClientFromEnv()
	if err != nil {
		return nil, maskAny(err)
	}

	return &DockerHost{
		Client:    client,
		IP:        hostIP,
		Endpoint:  endpoint,
		Interface: intf,
	}, nil

}

// getHostAddressForEndpoint returns the IP address of the host of the docker daemon with given endpoint.
func getHostAddressForEndpoint(endpoint, localHostIP string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", maskAny(err)
	}
	switch u.Scheme {
	case "http", "https", "tcp":
		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return "", maskAny(err)
		}
		return host, nil
	case "unix":
		return localHostIP, nil
	default:
		return "", maskAny(fmt.Errorf("Unsupported docker endpoint '%s'", endpoint))
	}
}
