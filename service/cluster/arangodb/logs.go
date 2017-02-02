package arangodb

import (
	"io"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/juju/errgo"
)

// CollectMachineLogs collects recent logs from the machine running the servers and writes them to the given writer.
func (m *arangodb) CollectMachineLogs(w io.Writer) error {
	// Collect logs from arangodb
	if err := m.collectLogs(w, m.containerID); err != nil && errgo.Cause(err) != io.EOF {
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
		if err := m.collectLogs(w, m.agentContainerID); err != nil && errgo.Cause(err) != io.EOF {
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
	if err := m.collectLogs(w, m.dbserverContainerID); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// CollectCoordinatorLogs collects recent logs from the coordinator and writes them to the given writer.
func (m *arangodb) CollectCoordinatorLogs(w io.Writer) error {
	if err := m.updateServerInfo(); err != nil {
		return maskAny(err)
	}
	if err := m.collectLogs(w, m.coordinatorContainerID); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// collectLogs collects recent logs from the container with given ID and writes them to the given writer.
func (m *arangodb) collectLogs(w io.Writer, containerID string) error {
	since := time.Now().Add(-time.Minute * 10)
	m.log.Debugf("fetching logs from %s", containerID)
	if err := m.dockerHost.client.Logs(docker.LogsOptions{
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
