package cluster

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

type arangodb struct {
	index           int
	ip              string
	port            int
	volumeID        string
	containerID     string
	hasAgent        bool
	agentPort       int
	coordinatorPort int
	dbserverPort    int
}

// launchArangodb creates a new container that runs arangodb.
// To do so, it creates a volume, then creates and starts the container.
func (c *arangodbCluster) launchArangodb(index int) (*arangodb, error) {
	// Create volume
	name := fmt.Sprintf("arangodb-%s-%d", c.id, index)
	c.log.Debugf("Creating docker volume for arangodb %d", index)
	vol, err := c.client.CreateVolume(docker.CreateVolumeOptions{
		Name: name + "-vol",
	})
	if err != nil {
		return nil, maskAny(err)
	}

	args := []string{
		fmt.Sprintf("--masterPort=%d", c.MasterPort),
		fmt.Sprintf("--dockerContainer=%s", name),
		fmt.Sprintf("--ownAddress=%s", c.ArangodbConfig.DockerHostIP),
	}
	if index > 0 {
		args = append(args,
			fmt.Sprintf("--join=%s:%d", c.ArangodbConfig.DockerHostIP, c.MasterPort),
		)
	}
	port := c.MasterPort + (index * 5)
	opts := docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Image:        c.ArangodbConfig.ArangodbImage,
			Cmd:          args,
			Tty:          true,
			ExposedPorts: make(map[docker.Port]struct{}),
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:%s", vol.Name, "/data"),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				docker.Port(fmt.Sprintf("%d/tcp", c.MasterPort)): []docker.PortBinding{
					docker.PortBinding{
						HostIP:   "0.0.0.0",
						HostPort: strconv.Itoa(port),
					},
				},
			},
			PublishAllPorts: true,
		},
	}
	if strings.HasPrefix(c.DockerEndpoint, "unix://") {
		path := strings.TrimPrefix(c.DockerEndpoint, "unix://")
		opts.HostConfig.Binds = append(opts.HostConfig.Binds,
			fmt.Sprintf("%s:%s", path, path),
		)
	}
	c.log.Debugf("Creating arangodb container %s", name)
	cont, err := c.client.CreateContainer(opts)
	if err != nil {
		return nil, maskAny(err)
	}
	c.log.Debugf("Starting arangodb container %s (%s)", name, cont.ID)
	if err := c.client.StartContainer(cont.ID, opts.HostConfig); err != nil {
		return nil, maskAny(err)
	}
	c.log.Debugf("Started arangodb container %s (%s)", name, cont.ID)
	return &arangodb{
		index:           index,
		ip:              c.ArangodbConfig.DockerHostIP,
		port:            port,
		volumeID:        vol.Name,
		containerID:     cont.ID,
		hasAgent:        index < c.agencySize,
		agentPort:       port + 1, // Ugly; this depends on knowledge of arangodb
		coordinatorPort: port + 2,
		dbserverPort:    port + 3,
	}, nil
}

// waitUntilServersReady blocks until all servers on the machine are ready.
func (m *arangodb) waitUntilServersReady(log *logging.Logger, timeout time.Duration) error {
	testInstance := func(address string, port int, name string, timeout time.Duration) error {
		log.Debugf("Waiting for %s-%d on %s:%d to get ready", name, m.index, address, port)
		start := time.Now()
		for {
			url := fmt.Sprintf("http://%s:%d/_api/version", address, port)
			r, e := http.Get(url)
			if e == nil && r != nil && r.StatusCode == 200 {
				log.Debugf("%s-%d on %s:%d is ready", name, m.index, address, port)
				return nil
			}

			if time.Since(start) > timeout {
				return maskAny(errgo.WithCausef(nil, TimeoutError, "%s-%d on %s:%d is not ready in time", name, m.index, address, port))
			}
			time.Sleep(time.Millisecond * 500)
		}
	}

	g := errgroup.Group{}
	if m.hasAgent {
		g.Go(func() error {
			return testInstance(m.ip, m.agentPort, "agent", timeout)
		})
	}
	g.Go(func() error {
		return testInstance(m.ip, m.coordinatorPort, "coordinator", timeout)
	})
	g.Go(func() error {
		return testInstance(m.ip, m.dbserverPort, "dbserver", timeout)
	})
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}
	return nil
}

func (m *arangodb) destroy() error {
	return nil
}
