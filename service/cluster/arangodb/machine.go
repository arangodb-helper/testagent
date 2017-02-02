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
	"sync/atomic"
	"time"

	"github.com/arangodb/testAgent/service/cluster"
	"github.com/cenkalti/backoff"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

const (
	stopMachineTimeout   = 120 // Seconds until a machine container that is stopped will be killed
	stopContainerTimeout = 45  // Seconds until a container that is stopped will be killed
	testStatusTimeout    = time.Second * 20
)

type arangodb struct {
	dockerHost                 *dockerHost
	log                        *logging.Logger
	createOptions              docker.CreateContainerOptions
	index                      int
	port                       int
	createdAt                  time.Time
	startedAt                  time.Time
	state                      cluster.MachineState
	volumeID                   string
	containerID                string
	hasAgent                   bool
	agentPort                  int
	agentContainerID           string
	lastAgentReadyStatus       int32
	coordinatorPort            int
	coordinatorContainerID     string
	lastCoordinatorReadyStatus int32
	dbserverPort               int
	dbserverContainerID        string
	lastDBServerReadyStatus    int32
}

// ID returns a unique identifier for this machine
func (m *arangodb) ID() string {
	return fmt.Sprintf("m%d-%s:%d", m.index, m.dockerHost.ip, m.coordinatorPort)
}

// State returns the current state of the machine
func (m *arangodb) State() cluster.MachineState {
	return m.state
}

// CreatedAt returns the time when this machine was created
func (m *arangodb) CreatedAt() time.Time {
	return m.createdAt
}

// StartedAt returns the time when this machine was last started
func (m *arangodb) StartedAt() time.Time {
	return m.startedAt
}

// HasAgent returns true if there is an agent on this machine
func (m *arangodb) HasAgent() bool {
	return m.hasAgent
}

// AgentURL returns the URL of the agent on this machine.
func (m *arangodb) AgentURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(m.dockerHost.ip, strconv.Itoa(m.agentPort)),
	}
}

// DBServerURL returns the URL of the DBServer on this machine.
func (m *arangodb) DBServerURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(m.dockerHost.ip, strconv.Itoa(m.dbserverPort)),
	}
}

// CoordinatorURL returns the URL of the Coordinator on this machine.
func (m *arangodb) CoordinatorURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(m.dockerHost.ip, strconv.Itoa(m.coordinatorPort)),
	}
}

// TestAgentStatus checks if the agent on this machine is ready (with a reasonable timeout). If returns nil on ready, error on not ready.
func (m *arangodb) TestAgentStatus() error {
	return maskAny(m.testInstance(nil, m.AgentURL(), "agent", testStatusTimeout, &m.lastAgentReadyStatus))
}

// TestDBServerStatus checks if the dbserver on this machine is ready (with a reasonable timeout). If returns nil on ready, error on not ready.
func (m *arangodb) TestDBServerStatus() error {
	return maskAny(m.testInstance(nil, m.DBServerURL(), "dbserver", testStatusTimeout, &m.lastDBServerReadyStatus))
}

// TestCoordinatorStatus checks if the coordinator on this machine is ready (with a reasonable timeout). If returns nil on ready, error on not ready.
func (m *arangodb) TestCoordinatorStatus() error {
	return maskAny(m.testInstance(nil, m.CoordinatorURL(), "coordinator", testStatusTimeout, &m.lastCoordinatorReadyStatus))
}

// LastAgentReadyStatus returns true if the last known agent ready check succeeded.
func (m *arangodb) LastAgentReadyStatus() bool {
	return m.lastAgentReadyStatus != 0
}

// LastDBServerReadyStatus returns true if the last known dbserver ready check succeeded.
func (m *arangodb) LastDBServerReadyStatus() bool {
	return m.lastDBServerReadyStatus != 0
}

// LastCoordinatorReadyStatus returns true if the last known coordinator ready check succeeded.
func (m *arangodb) LastCoordinatorReadyStatus() bool {
	return m.lastCoordinatorReadyStatus != 0
}

// Perform a graceful restart of the agent. This function does NOT wait until the agent is ready again.
func (m *arangodb) RestartAgent() error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.dockerHost.client.StopContainer(m.agentContainerID, stopContainerTimeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Perform a graceful restart of the dbserver. This function does NOT wait until the dbserver is ready again.
func (m *arangodb) RestartDBServer() error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.dockerHost.client.StopContainer(m.dbserverContainerID, stopContainerTimeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Perform a graceful restart of the coordinator. This function does NOT wait until the coordinator is ready again.
func (m *arangodb) RestartCoordinator() error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.dockerHost.client.StopContainer(m.coordinatorContainerID, stopContainerTimeout); err != nil {
		return maskAny(err)
	}
	return nil
}

// Perform a forced restart of the agent. This function does NOT wait until the agent is ready again.
func (m *arangodb) KillAgent() error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.dockerHost.client.KillContainer(docker.KillContainerOptions{ID: m.agentContainerID}); err != nil {
		return maskAny(err)
	}
	return nil
}

// Perform a forced restart of the dbserver. This function does NOT wait until the dbserver is ready again.
func (m *arangodb) KillDBServer() error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.dockerHost.client.KillContainer(docker.KillContainerOptions{ID: m.dbserverContainerID}); err != nil {
		return maskAny(err)
	}
	return nil
}

// Perform a forced restart of the coordinator. This function does NOT wait until the coordinator is ready again.
func (m *arangodb) KillCoordinator() error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.dockerHost.client.KillContainer(docker.KillContainerOptions{ID: m.coordinatorContainerID}); err != nil {
		return maskAny(err)
	}
	return nil
}

// Reboot performs a graceful reboot of the machine
func (m *arangodb) Reboot() error {
	// Stop the arangodb container  (it will stop the servers )
	m.log.Infof("'Rebooting' container %s", m.containerID)
	if err := m.dockerHost.client.StopContainer(m.containerID, stopMachineTimeout); err != nil {
		return maskAny(err)
	}

	// Remove container
	m.log.Infof("Removing container %s", m.containerID)
	if err := m.dockerHost.client.RemoveContainer(docker.RemoveContainerOptions{
		Force: true, // Just in case
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

// DestroyAllowed returns true if it is allowed to destroy this machine
func (m *arangodb) DestroyAllowed() bool {
	return m.index > 0
}

// Remove the machine without the ability to recover it
func (m *arangodb) Destroy() error {
	// Terminate arangodb. It will terminate the servers.
	if err := m.stop(); err != nil {
		return maskAny(err)
	}

	// Set state
	m.state = cluster.MachineStateDestroyed

	// Remove volume
	m.log.Infof("Removing volume %s", m.volumeID)
	if err := m.dockerHost.client.RemoveVolume(m.volumeID); err != nil {
		return maskAny(err)
	}
	return nil
}

// createMachine creates a volume and all configuration needed to start arangodb.
func (c *arangodbCluster) createMachine(index int) (*arangodb, error) {
	// Pick a docker host
	dockerHost := c.dockerHosts[index%len(c.dockerHosts)]

	// Create volume
	name := fmt.Sprintf("arangodb-%s-%d", c.id, index)
	volName := name + "-vol"
	c.log.Debugf("Creating docker volume for arangodb %d on %s", index, dockerHost.ip)
	_, err := dockerHost.client.CreateVolume(docker.CreateVolumeOptions{
		Name: volName,
	})
	if err != nil {
		return nil, maskAny(err)
	}

	args := []string{
		fmt.Sprintf("--masterPort=%d", c.MasterPort),
		fmt.Sprintf("--dockerContainer=%s", name),
		fmt.Sprintf("--dockerEndpoint=%s", dockerHost.endpoint),
		fmt.Sprintf("--ownAddress=%s", dockerHost.ip),
	}
	if c.Verbose {
		args = append(args, "--verbose")
	}
	if c.DockerNetHost {
		args = append(args, "--dockerNetHost")
	}
	if c.ArangoImage != "" {
		args = append(args,
			fmt.Sprintf("--docker=%s", c.ArangodbConfig.ArangoImage),
		)
	}
	if index > 0 {
		args = append(args,
			fmt.Sprintf("--join=%s:%d", c.dockerHosts[0].ip, c.MasterPort),
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
			PublishAllPorts: false,
		},
	}
	if c.DockerNetHost {
		opts.HostConfig.NetworkMode = "host"
	}
	if strings.HasPrefix(dockerHost.endpoint, "unix://") {
		path := strings.TrimPrefix(dockerHost.endpoint, "unix://")
		opts.HostConfig.Binds = append(opts.HostConfig.Binds,
			fmt.Sprintf("%s:%s", path, path),
		)
	}
	return &arangodb{
		dockerHost:    dockerHost,
		createOptions: opts,
		log:           c.log,
		index:         index,
		createdAt:     time.Now(),
		state:         cluster.MachineStateNew,
		port:          port,
		volumeID:      volName,
	}, nil
}

// pullImage pulls a docker image on the docker host.
func (m *arangodb) pullImageIfNeeded(image string) error {
	if _, err := m.dockerHost.client.InspectImage(image); err == nil {
		// Image already available, do nothing
		return nil
	}
	repo, tag := docker.ParseRepositoryTag(image)
	m.log.Infof("Pulling %s:%s", repo, tag)
	if err := m.dockerHost.client.PullImage(docker.PullImageOptions{
		Repository: repo,
		Tag:        tag,
	}, docker.AuthConfiguration{}); err != nil {
		return maskAny(err)
	}
	return nil
}

// start the machine
func (m *arangodb) start() error {
	m.startedAt = time.Now()
	m.log.Debugf("Creating arangodb container %s", m.createOptions.Name)
	cont, err := m.dockerHost.client.CreateContainer(m.createOptions)
	if err != nil {
		return maskAny(err)
	}
	m.containerID = cont.ID
	m.log.Debugf("Starting arangodb container %s (%s)", m.createOptions.Name, cont.ID)
	if err := m.dockerHost.client.StartContainer(cont.ID, m.createOptions.HostConfig); err != nil {
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
	if err := m.dockerHost.client.StopContainer(m.containerID, stopMachineTimeout); err != nil {
		return maskAny(err)
	}

	// Remove container
	m.log.Infof("Removing container %s", m.containerID)
	if err := m.dockerHost.client.RemoveContainer(docker.RemoveContainerOptions{
		Force: true, // Just in case
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
	addr := fmt.Sprintf("http://%s:%d/process", m.dockerHost.ip, m.port)
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

	g := errgroup.Group{}
	if m.hasAgent {
		g.Go(func() error {
			return m.testInstance(m.log, m.AgentURL(), "agent", timeout, &m.lastAgentReadyStatus)
		})
	}
	g.Go(func() error {
		return m.testInstance(m.log, m.CoordinatorURL(), "coordinator", timeout, &m.lastCoordinatorReadyStatus)
	})
	g.Go(func() error {
		return m.testInstance(m.log, m.DBServerURL(), "dbserver", timeout, &m.lastDBServerReadyStatus)
	})
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}
	m.state = cluster.MachineStateReady
	return nil
}

// Test all servers to be up
func (m *arangodb) testInstance(log *logging.Logger, url url.URL, name string, timeout time.Duration, activeVar *int32) error {
	if log != nil {
		log.Debugf("Waiting for %s-%d on %s to get ready", name, m.index, url.String())
	}
	start := time.Now()
	client := &http.Client{Timeout: time.Second * 5}
	for {
		versionURL := url
		versionURL.Path = "/_api/version"
		r, e := client.Get(versionURL.String())
		if e == nil && r != nil && r.StatusCode == 200 {
			atomic.StoreInt32(activeVar, 1)
			if log != nil {
				log.Debugf("%s-%d on %s is ready", name, m.index, url.String())
			}
			return nil
		}

		atomic.StoreInt32(activeVar, 0)
		if time.Since(start) > timeout {
			return maskAny(errgo.WithCausef(nil, cluster.TimeoutError, "%s-%d on %s is not ready in time", name, m.index, url.String()))
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// watchdog monitors all servers and updates the last ready flag.
func (m *arangodb) watchdog() {
	timeout := time.Minute
	monitorLoop := func(urlGetter func() url.URL, name string, activeVar *int32) {
		for {
			switch m.state {
			case cluster.MachineStateReady:
				m.testInstance(nil, urlGetter(), name, timeout, activeVar)
			case cluster.MachineStateDestroyed:
				return // We're done
			}
			time.Sleep(time.Second * 15)
		}
	}
	if m.HasAgent() {
		go monitorLoop(func() url.URL { return m.AgentURL() }, "agent", &m.lastAgentReadyStatus)
	}
	go monitorLoop(func() url.URL { return m.DBServerURL() }, "dbserver", &m.lastDBServerReadyStatus)
	go monitorLoop(func() url.URL { return m.CoordinatorURL() }, "coordinator", &m.lastCoordinatorReadyStatus)
}
