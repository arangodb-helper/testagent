package arangodb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arangodb/testAgent/service/cluster"
	"github.com/cenkalti/backoff"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

const (
	stopMachineTimeout = 120 // Seconds until a machine container that is stopped will be killed
)

type arangodb struct {
	client                 *docker.Client
	log                    *logging.Logger
	createOptions          docker.CreateContainerOptions
	index                  int
	ip                     string
	port                   int
	state                  cluster.MachineState
	volumeID               string
	containerID            string
	hasAgent               bool
	agentPort              int
	agentContainerID       string
	coordinatorPort        int
	coordinatorContainerID string
	dbserverPort           int
	dbserverContainerID    string
}

// State returns the current state of the machine
func (m *arangodb) State() cluster.MachineState {
	return m.state
}

// HasAgent returns true if there is an agent on this machine
func (m *arangodb) HasAgent() bool {
	return m.hasAgent
}

// AgentURL returns the URL of the agent on this machine.
func (m *arangodb) AgentURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(m.ip, strconv.Itoa(m.agentPort)),
	}
}

// DBServerURL returns the URL of the DBServer on this machine.
func (m *arangodb) DBServerURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(m.ip, strconv.Itoa(m.dbserverPort)),
	}
}

// CoordinatorURL returns the URL of the Coordinator on this machine.
func (m *arangodb) CoordinatorURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(m.ip, strconv.Itoa(m.coordinatorPort)),
	}
}

// Reboot performs a graceful reboot of the machine
func (m *arangodb) Reboot() error {
	// Stop the arangodb container  (it will stop the servers )
	m.log.Infof("'Rebooting' container %s", m.containerID)
	if err := m.client.StopContainer(m.containerID, stopMachineTimeout); err != nil {
		return maskAny(err)
	}

	// Remove container
	m.log.Infof("Removing container %s", m.containerID)
	if err := m.client.RemoveContainer(docker.RemoveContainerOptions{
		Force: true, // Jut in case
		ID:    m.containerID,
	}); err != nil {
		return maskAny(err)
	}

	// Relaunch
	if err := m.start(); err != nil {
		return maskAny(err)
	}

	// Wait for servers ready
	if err := m.waitUntilServersReady(m.log, serverReadyTimeout); err != nil {
		return maskAny(err)
	}

	return nil
}

// Remove the machine without the ability to recover it
func (m *arangodb) Destroy() error {
	// Terminate arangodb. It will terminate the servers.
	if err := m.stop(); err != nil {
		return maskAny(err)
	}

	// Remove volume
	m.log.Infof("Removing volume %s", m.volumeID)
	if err := m.client.RemoveVolume(m.volumeID); err != nil {
		return maskAny(err)
	}
	return nil
}

// createMachine creates a volume and all configuration needed to start arangodb.
func (c *arangodbCluster) createMachine(index int) (*arangodb, error) {
	// Create volume
	name := fmt.Sprintf("arangodb-%s-%d", c.id, index)
	volName := name + "-vol"
	c.log.Debugf("Creating docker volume for arangodb %d", index)
	_, err := c.client.CreateVolume(docker.CreateVolumeOptions{
		Name: volName,
	})
	if err != nil {
		return nil, maskAny(err)
	}

	args := []string{
		fmt.Sprintf("--masterPort=%d", c.MasterPort),
		fmt.Sprintf("--dockerContainer=%s", name),
		fmt.Sprintf("--ownAddress=%s", c.ArangodbConfig.DockerHostIP),
	}
	if c.ArangoImage != "" {
		args = append(args,
			fmt.Sprintf("--docker=%s", c.ArangodbConfig.ArangoImage),
		)
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
			Image: c.ArangodbConfig.ArangodbImage,
			Cmd:   args,
			Tty:   true,
			ExposedPorts: map[docker.Port]struct{}{
				docker.Port(fmt.Sprintf("%d/tcp", c.MasterPort)): struct{}{},
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:%s", volName, "/data"),
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
	return &arangodb{
		client:        c.client,
		createOptions: opts,
		log:           c.log,
		index:         index,
		ip:            c.ArangodbConfig.DockerHostIP,
		state:         cluster.MachineStateNew,
		port:          port,
		volumeID:      volName,
	}, nil
}

// start the machine
func (m *arangodb) start() error {
	m.log.Debugf("Creating arangodb container %s", m.createOptions.Name)
	cont, err := m.client.CreateContainer(m.createOptions)
	if err != nil {
		return maskAny(err)
	}
	m.containerID = cont.ID
	m.log.Debugf("Starting arangodb container %s (%s)", m.createOptions.Name, cont.ID)
	if err := m.client.StartContainer(cont.ID, m.createOptions.HostConfig); err != nil {
		return maskAny(err)
	}
	m.log.Debugf("Started arangodb container %s (%s)", m.createOptions.Name, cont.ID)
	m.state = cluster.MachineStateStarted
	return nil
}

func (m *arangodb) stop() error {
	// Stop the arangodb container  (it will stop the servers )
	m.state = cluster.MachineStateShutdown
	m.log.Infof("Stopping container %s", m.containerID)
	if err := m.client.StopContainer(m.containerID, stopMachineTimeout); err != nil {
		return maskAny(err)
	}

	// Remove container
	m.log.Infof("Removing container %s", m.containerID)
	if err := m.client.RemoveContainer(docker.RemoveContainerOptions{
		Force: true, // Jut in case
		ID:    m.containerID,
	}); err != nil {
		return maskAny(err)
	}
	return nil
}

type ProcessListResponse struct {
	ServersStarted bool            `json:"servers-started,omitempty"` // True if the server have all been started
	Servers        []ServerProcess `json:"servers,omitempty"`         // List of servers started by ArangoDB
}

type ServerProcess struct {
	Type        string `json:"type"`                   // agent | coordinator | dbserver
	IP          string `json:"ip"`                     // IP address needed to reach the server
	Port        int    `json:"port"`                   // Port needed to reach the server
	ProcessID   int    `json:"pid,omitempty"`          // PID of the process (0 when running in docker)
	ContainerID string `json:"container-id,omitempty"` // ID of docker container running the server
}

// updateServerInfo connects to arangodb to query the port numbers & container info
// of all servers on the machine
func (m *arangodb) updateServerInfo() error {
	addr := fmt.Sprintf("http://%s:%d/process", m.ip, m.port)
	fetchInfo := func() error {
		resp, err := http.Get(addr)
		if err != nil {
			return maskAny(err)
		}
		if resp.StatusCode != http.StatusOK {
			return maskAny(fmt.Errorf("Invalid status; expected %d, got %d", http.StatusOK, resp.StatusCode))
		}
		var plResp ProcessListResponse
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return maskAny(err)
		}
		if err := json.Unmarshal(body, &plResp); err != nil {
			return maskAny(err)
		}
		if !plResp.ServersStarted {
			return maskAny(fmt.Errorf("Servers not yet started"))
		}
		hasAgent := false
		for _, s := range plResp.Servers {
			switch s.Type {
			case "agent":
				m.agentPort = s.Port
				m.agentContainerID = s.ContainerID
				hasAgent = true
			case "coordinator":
				m.coordinatorPort = s.Port
				m.coordinatorContainerID = s.ContainerID
			case "dbserver":
				m.dbserverPort = s.Port
				m.dbserverContainerID = s.ContainerID
			}
		}
		m.hasAgent = hasAgent
		return nil
	}

	err := backoff.Retry(fetchInfo, backoff.NewExponentialBackOff())
	if err != nil {
		return maskAny(err)
	}

	return nil
}

// waitUntilServersReady blocks until all servers on the machine are ready.
func (m *arangodb) waitUntilServersReady(log *logging.Logger, timeout time.Duration) error {
	// First wait for all servers to be started and we know their addresses
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}

	// Test all servers to be up
	testInstance := func(url url.URL, name string, timeout time.Duration) error {
		log.Debugf("Waiting for %s-%d on %s to get ready", name, m.index, url.String())
		start := time.Now()
		for {
			versionURL := url
			versionURL.Path = "/_api/version"
			r, e := http.Get(versionURL.String())
			if e == nil && r != nil && r.StatusCode == 200 {
				log.Debugf("%s-%d on %s is ready", name, m.index, url.String())
				return nil
			}

			if time.Since(start) > timeout {
				return maskAny(errgo.WithCausef(nil, cluster.TimeoutError, "%s-%d on %s is not ready in time", name, m.index, url.String()))
			}
			time.Sleep(time.Millisecond * 500)
		}
	}

	g := errgroup.Group{}
	if m.hasAgent {
		g.Go(func() error {
			return testInstance(m.AgentURL(), "agent", timeout)
		})
	}
	g.Go(func() error {
		return testInstance(m.CoordinatorURL(), "coordinator", timeout)
	})
	g.Go(func() error {
		return testInstance(m.DBServerURL(), "dbserver", timeout)
	})
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}
	m.state = cluster.MachineStateReady
	return nil
}
