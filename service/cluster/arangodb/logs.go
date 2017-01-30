package arangodb

import (
	"io"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

// CollectAgentLogs collects recent logs from the agent and writes them to the given writer.
func (m *arangodb) CollectAgentLogs(w io.Writer) error {
	if m.HasAgent() {
		since := time.Now().Add(-time.Minute * 10)
		err := m.client.Logs(docker.LogsOptions{
			Container:    m.agentContainerID,
			OutputStream: w,
			ErrorStream:  w,
			Stdout:       true,
			Stderr:       true,
			Since:        since.Unix(),
			Timestamps:   true,
		})
		return maskAny(err)
	}
	return nil
}

// CollectDBServerLogs collects recent logs from the dbserver and writes them to the given writer.
func (m *arangodb) CollectDBServerLogs(w io.Writer) error {
	since := time.Now().Add(-time.Minute * 10)
	err := m.client.Logs(docker.LogsOptions{
		Container:    m.dbserverContainerID,
		OutputStream: w,
		ErrorStream:  w,
		Stdout:       true,
		Stderr:       true,
		Since:        since.Unix(),
		Timestamps:   true,
	})
	return maskAny(err)
}

// CollectCoordinatorLogs collects recent logs from the coordinator and writes them to the given writer.
func (m *arangodb) CollectCoordinatorLogs(w io.Writer) error {
	since := time.Now().Add(-time.Minute * 10)
	err := m.client.Logs(docker.LogsOptions{
		Container:    m.coordinatorContainerID,
		OutputStream: w,
		ErrorStream:  w,
		Stdout:       true,
		Stderr:       true,
		Since:        since.Unix(),
		Timestamps:   true,
	})
	return maskAny(err)
}
