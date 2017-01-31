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
	since := time.Now().Add(-time.Minute * 10)
	if err := m.client.Logs(docker.LogsOptions{
		Container:    m.containerID,
		OutputStream: w,
		ErrorStream:  w,
		Stdout:       true,
		Stderr:       true,
		Since:        since.Unix(),
		Timestamps:   true,
	}); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// CollectAgentLogs collects recent logs from the agent and writes them to the given writer.
func (m *arangodb) CollectAgentLogs(w io.Writer) error {
	if m.HasAgent() {
		since := time.Now().Add(-time.Minute * 10)
		if err := m.client.Logs(docker.LogsOptions{
			Container:    m.agentContainerID,
			OutputStream: w,
			ErrorStream:  w,
			Stdout:       true,
			Stderr:       true,
			Since:        since.Unix(),
			Timestamps:   true,
		}); err != nil && errgo.Cause(err) != io.EOF {
			return maskAny(err)
		}
		return nil
	}
	return nil
}

// CollectDBServerLogs collects recent logs from the dbserver and writes them to the given writer.
func (m *arangodb) CollectDBServerLogs(w io.Writer) error {
	since := time.Now().Add(-time.Minute * 10)
	if err := m.client.Logs(docker.LogsOptions{
		Container:    m.dbserverContainerID,
		OutputStream: w,
		ErrorStream:  w,
		Stdout:       true,
		Stderr:       true,
		Since:        since.Unix(),
		Timestamps:   true,
	}); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}

// CollectCoordinatorLogs collects recent logs from the coordinator and writes them to the given writer.
func (m *arangodb) CollectCoordinatorLogs(w io.Writer) error {
	since := time.Now().Add(-time.Minute * 10)
	if err := m.client.Logs(docker.LogsOptions{
		Container:    m.coordinatorContainerID,
		OutputStream: w,
		ErrorStream:  w,
		Stdout:       true,
		Stderr:       true,
		Since:        since.Unix(),
		Timestamps:   true,
	}); err != nil && errgo.Cause(err) != io.EOF {
		return maskAny(err)
	}
	return nil
}
