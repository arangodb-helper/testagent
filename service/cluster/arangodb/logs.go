package arangodb

import (
	"fmt"
	"io"
	"net/http"
	"time"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/juju/errgo"
)

// CollectMachineLogs collects recent logs from the machine running the servers and writes them to the given writer.
func (m *arangodb) CollectMachineLogs(w io.Writer) error {
	// Collect logs from arangodb
	if err := m.collectContainerLogs(w, m.containerID); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// CollectNetworkLogs collects recent logs from the network(-blocker) running the servers and writes them to the given writer.
func (m *arangodb) CollectNetworkLogs(w io.Writer) error {
	// Collect logs from network-blocker
	if err := m.collectContainerLogs(w, m.nwBlockerContainerID); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// CollectAgentLogs collects recent logs from the agent and writes them to the given writer.
func (m *arangodb) CollectAgentLogs(w io.Writer) error {
	if m.HasAgent() {
		if err := m.updateServerInfo(); err != nil {
			return maskAny(err)
		}
		if err := m.collectServerLogs(w, "agent"); err != nil && errgo.Cause(err) != io.EOF {
			return maskAny(err)
		}
		return nil
	}
	return nil
}

// CollectDBServerLogs collects recent logs from the dbserver and writes them to the given writer.
func (m *arangodb) CollectDBServerLogs(w io.Writer) error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.collectServerLogs(w, "dbserver"); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// CollectCoordinatorLogs collects recent logs from the coordinator and writes them to the given writer.
func (m *arangodb) CollectCoordinatorLogs(w io.Writer) error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.collectServerLogs(w, "coordinator"); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// collectContainerLogs collects recent logs from the container with given ID and writes them to the given writer.
func (m *arangodb) collectContainerLogs(w io.Writer, containerID string) error {
	since := time.Now().Add(-time.Minute * 10)
	m.log.Debugf("fetching logs from %s", containerID)
	if err := m.dockerHost.Client.Logs(dc.LogsOptions{
		Container:    containerID,
		OutputStream: w,
		RawTerminal:  true,
		Stdout:       true,
		Stderr:       true,
		Since:        since.Unix(),
		Timestamps:   false,
	}); err != nil && errgo.Cause(err) != io.EOF {
		m.log.Debugf("failed to fetching logs from %s: %v", containerID, err)
		return maskAny(err)
	}
	m.log.Debugf("done fetching logs from %s", containerID)
	return nil
}

// collectServerLogs collects recent logs from the container with given ID and writes them to the given writer.
func (m *arangodb) collectServerLogs(w io.Writer, server string) error {
	addr := fmt.Sprintf("http://%s:%d/logs/%s", m.dockerHost.IP, m.arangodbPort, server)
	m.log.Debugf("fetching logs from %s", addr)

	resp, err := http.Get(addr)
	if err != nil {
		m.log.Debugf("failed to fetching logs from %s: %v", addr, err)
		return maskAny(err)
	}
	if resp.StatusCode != http.StatusOK {
		return maskAny(fmt.Errorf("Invalid status; expected %d, got %d", http.StatusOK, resp.StatusCode))
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return maskAny(err)
	}
	return nil
}
