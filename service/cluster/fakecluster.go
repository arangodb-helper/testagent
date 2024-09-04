package cluster

import (
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"
)

type FakeClusterBuilder struct {
	NrCoordinators int // ports 8530, ...
	NrDBServers    int // ports 8629, ...
	NrAgents       int // ports 4001, ...
}

func NewFakeCluster(nrCoord, nrDBS, nrAg int) *FakeClusterBuilder {
	return &FakeClusterBuilder{
		NrCoordinators: nrCoord,
		NrDBServers:    nrDBS,
		NrAgents:       nrAg,
	}
}

type FakeMachine struct {
	id    string
	index int
	fc    *FakeCluster
}

func (m *FakeMachine) ID() string {
	return m.id
}

func (m *FakeMachine) State() MachineState {
	return MachineStateReady
}

func (m *FakeMachine) CreatedAt() time.Time {
	return time.Now().Add(time.Hour * -7 * 24)
}

func (m *FakeMachine) StartedAt() time.Time {
	return time.Now().Add(time.Hour * -7 * 24)
}

func (m *FakeMachine) HasAgent() bool {
	return m.index < m.fc.fcb.NrAgents
}

func (m *FakeMachine) AgentURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(4001+m.index)),
	}
}

func (m *FakeMachine) DBServerURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(8629+m.index)),
	}
}

func (m *FakeMachine) CoordinatorURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(8530+m.index)),
	}
}

func (m *FakeMachine) LastAgentReadyStatus() bool {
	return true
}

func (m *FakeMachine) LastDBServerReadyStatus() bool {
	return true
}

func (m *FakeMachine) LastCoordinatorReadyStatus() bool {
	return true
}

func (m *FakeMachine) TestAgentStatus() error {
	return nil
}

func (m *FakeMachine) TestDBServerStatus() error {
	return nil
}

func (m *FakeMachine) TestCoordinatorStatus() error {
	return nil
}

func (m *FakeMachine) RestartAgent() error {
	return nil
}

func (m *FakeMachine) RestartDBServer() error {
	return nil
}

func (m *FakeMachine) RestartCoordinator() error {
	return nil
}

func (m *FakeMachine) KillAgent() error {
	return nil
}

func (m *FakeMachine) KillDBServer() error {
	return nil
}

func (m *FakeMachine) KillCoordinator() error {
	return nil
}

func (m *FakeMachine) RejectAgentTraffic() error {
	return nil
}

func (m *FakeMachine) RejectDBServerTraffic() error {
	return nil
}

func (m *FakeMachine) RejectCoordinatorTraffic() error {
	return nil
}

func (m *FakeMachine) DropAgentTraffic() error {
	return nil
}

func (m *FakeMachine) DropDBServerTraffic() error {
	return nil
}

func (m *FakeMachine) DropCoordinatorTraffic() error {
	return nil
}

func (m *FakeMachine) AcceptAgentTraffic() error {
	return nil
}

func (m *FakeMachine) AcceptDBServerTraffic() error {
	return nil
}

func (m *FakeMachine) AcceptCoordinatorTraffic() error {
	return nil
}

func (m *FakeMachine) CollectMachineLogs(w io.Writer) error {
	_, err := w.Write([]byte("FakeLog\n"))
	return err
}

func (m *FakeMachine) CollectNetworkLogs(w io.Writer) error {
	_, err := w.Write([]byte("FakeLog\n"))
	return err
}

func (m *FakeMachine) CollectAgentLogs(w io.Writer) error {
	_, err := w.Write([]byte("FakeLog\n"))
	return err
}

func (m *FakeMachine) CollectDBServerLogs(w io.Writer) error {
	_, err := w.Write([]byte("FakeLog\n"))
	return err
}

func (m *FakeMachine) CollectCoordinatorLogs(w io.Writer) error {
	_, err := w.Write([]byte("FakeLog\n"))
	return err
}

func (m *FakeMachine) CollectNetworkRules() ([]string, error) {
	return []string{}, nil
}

func (m *FakeMachine) Reboot() error {
	return nil
}

func (m *FakeMachine) DestroyAllowed() bool {
	return false
}

func (m *FakeMachine) Destroy() error {
	return errors.New("Cannot destroy")
}

type FakeCluster struct {
	id  string
	fcb *FakeClusterBuilder
}

// Now need to implement a FakeCluster:

func (fc *FakeCluster) ID() string {
	return fc.id
}

func (fc *FakeCluster) Add() (Machine, error) {
	return nil, errors.New("Cannot add machines to fake clusters")
}

func (fc *FakeCluster) ArangoImage() string {
	return "none"
}

func (fc *FakeCluster) WaitUntilReady() error {
	return nil
}

func (fc *FakeCluster) Destroy() error {
	return nil
}

func (fc *FakeCluster) StartMetricsCollection() error {
	return nil
}

func (fc *FakeCluster) Machines() ([]Machine, error) {
	list := make([]Machine, 0, 3)
	var i int = 0
	for {
		list = append(list, &FakeMachine{
			id:    "m" + strconv.Itoa(i),
			index: i,
			fc:    fc,
		})
		i += 1
		if i >= fc.fcb.NrAgents && i >= fc.fcb.NrDBServers && i >= fc.fcb.NrCoordinators {
			break
		}
	}
	return list, nil
}

func (fcb *FakeClusterBuilder) Create(agencySize int, forceOneShard bool) (Cluster, error) {
	if agencySize == fcb.NrAgents {
		return &FakeCluster{
			id:  "abc",
			fcb: fcb,
		}, nil
	} else {
		return nil, errors.New("Wrong number of agents")
	}
}
