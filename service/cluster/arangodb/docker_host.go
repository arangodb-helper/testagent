package arangodb

import docker "github.com/fsouza/go-dockerclient"

type dockerHost struct {
	client   *docker.Client
	ip       string
	endpoint string
}

func newDockerHost(endpoint, hostIP string) (*dockerHost, error) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return nil, maskAny(err)
	}
	return &dockerHost{
		client:   client,
		ip:       hostIP,
		endpoint: endpoint,
	}, nil
}
