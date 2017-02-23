package arangodb

import (
	"fmt"

	"github.com/pkg/errors"
)

// Actively reject all network traffic to the agent
func (m *arangodb) RejectAgentTraffic() error {
	if !m.HasAgent() {
		return maskAny(fmt.Errorf("no agent on this machine"))
	}
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.agentContainerIP == "" {
		return maskAny(fmt.Errorf("agent container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.RejectTCP(m.agentPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to reject agent traffic (to)"))
		}
		if err := api.RejectAllFrom(m.agentContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to reject agent traffic (from)"))
		}
	}
	return nil
}

// Actively reject all network traffic to the dbserver
func (m *arangodb) RejectDBServerTraffic() error {
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.dbserverContainerIP == "" {
		return maskAny(fmt.Errorf("dbserver container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.RejectTCP(m.dbserverPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to reject dbserver traffic (to)"))
		}
		if err := api.RejectAllFrom(m.dbserverContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to reject dbserver traffic (from)"))
		}
	}
	return nil
}

// Actively reject all network traffic to the coordinator
func (m *arangodb) RejectCoordinatorTraffic() error {
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.coordinatorContainerIP == "" {
		return maskAny(fmt.Errorf("coordinator container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.RejectTCP(m.coordinatorPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to reject coordinator traffic (to)"))
		}
		if err := api.RejectAllFrom(m.coordinatorContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to reject coordinator traffic (from)"))
		}
	}
	return nil
}

// Silently drop all network traffic to the agent
func (m *arangodb) DropAgentTraffic() error {
	if !m.HasAgent() {
		return maskAny(fmt.Errorf("no agent on this machine"))
	}
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.agentContainerIP == "" {
		return maskAny(fmt.Errorf("agent container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.DropTCP(m.agentPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to drop agent traffic (to)"))
		}
		if err := api.DropAllFrom(m.agentContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to drop agent traffic (from)"))
		}
	}
	return nil
}

// Silently drop all network traffic to the dbserver
func (m *arangodb) DropDBServerTraffic() error {
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.dbserverContainerIP == "" {
		return maskAny(fmt.Errorf("dbserver container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.DropTCP(m.dbserverPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to drop dbserver traffic (to)"))
		}
		if err := api.DropAllFrom(m.dbserverContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to drop dbserver traffic (from)"))
		}
	}
	return nil
}

// Silently drop all network traffic to the coordinator
func (m *arangodb) DropCoordinatorTraffic() error {
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.coordinatorContainerIP == "" {
		return maskAny(fmt.Errorf("coordinator container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.DropTCP(m.coordinatorPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to drop coordinator traffic (to)"))
		}
		if err := api.DropAllFrom(m.coordinatorContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to drop coordinator traffic (from)"))
		}
	}
	return nil
}

// Accept all network traffic to the agent
func (m *arangodb) AcceptAgentTraffic() error {
	if !m.HasAgent() {
		return maskAny(fmt.Errorf("no agent on this machine"))
	}
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.agentContainerIP == "" {
		return maskAny(fmt.Errorf("agent container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.AcceptTCP(m.agentPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to accept agent traffic (to)"))
		}
		if err := api.AcceptAllFrom(m.agentContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to accept agent traffic (from)"))
		}
	}
	return nil
}

// Accept all network traffic to the dbserver
func (m *arangodb) AcceptDBServerTraffic() error {
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.dbserverContainerIP == "" {
		return maskAny(fmt.Errorf("dbserver container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.AcceptTCP(m.dbserverPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to accept dbserver traffic (to)"))
		}
		if err := api.AcceptAllFrom(m.dbserverContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to accept dbserver traffic (from)"))
		}
	}
	return nil
}

// Accept all network traffic to the coordinator
func (m *arangodb) AcceptCoordinatorTraffic() error {
	if m.createOptions.HostConfig.NetworkMode == "host" {
		return maskAny(fmt.Errorf("network operations are nt supported on host networking"))
	}
	if m.coordinatorContainerIP == "" {
		return maskAny(fmt.Errorf("coordinator container IP is unknown"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.AcceptTCP(m.coordinatorPort); err != nil {
			return maskAny(errors.Wrap(err, "Failed to accept coordinator traffic (to)"))
		}
		if err := api.AcceptAllFrom(m.coordinatorContainerIP, m.dockerHost.Interface); err != nil {
			return maskAny(errors.Wrap(err, "Failed to accept coordinator traffic (from)"))
		}
	}
	return nil
}

// CollectNetworkRules fetches all network rules that are involve one of the servers
func (m *arangodb) CollectNetworkRules() ([]string, error) {
	if api := m.nwBlocker; api == nil {
		return nil, maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if list, err := api.Rules(); err != nil {
			return nil, maskAny(errors.Wrap(err, "Failed to list network rules"))
		} else {
			return list, nil
		}
	}
}
