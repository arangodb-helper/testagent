package arangodb

import "fmt"
import "github.com/juju/errgo"

// Actively reject all network traffic to the agent
func (m *arangodb) RejectAgentTraffic() error {
	if !m.HasAgent() {
		return maskAny(fmt.Errorf("no agent on this machine"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.RejectTCP(m.agentPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to reject agent traffic"))
		}
	}
	return nil
}

// Actively reject all network traffic to the dbserver
func (m *arangodb) RejectDBServerTraffic() error {
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.RejectTCP(m.dbserverPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to reject dbserver traffic"))
		}
	}
	return nil
}

// Actively reject all network traffic to the coordinator
func (m *arangodb) RejectCoordinatorTraffic() error {
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.RejectTCP(m.coordinatorPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to reject coordinator traffic"))
		}
	}
	return nil
}

// Silently drop all network traffic to the agent
func (m *arangodb) DropAgentTraffic() error {
	if !m.HasAgent() {
		return maskAny(fmt.Errorf("no agent on this machine"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.DropTCP(m.agentPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to drop agent traffic"))
		}
	}
	return nil
}

// Silently drop all network traffic to the dbserver
func (m *arangodb) DropDBServerTraffic() error {
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.DropTCP(m.dbserverPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to drop dbserver traffic"))
		}
	}
	return nil
}

// Silently drop all network traffic to the coordinator
func (m *arangodb) DropCoordinatorTraffic() error {
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.DropTCP(m.coordinatorPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to drop coordinator traffic"))
		}
	}
	return nil
}

// Accept all network traffic to the agent
func (m *arangodb) AcceptAgentTraffic() error {
	if !m.HasAgent() {
		return maskAny(fmt.Errorf("no agent on this machine"))
	}
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.AcceptTCP(m.agentPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to accept agent traffic"))
		}
	}
	return nil
}

// Accept all network traffic to the dbserver
func (m *arangodb) AcceptDBServerTraffic() error {
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.AcceptTCP(m.dbserverPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to accept dbserver traffic"))
		}
	}
	return nil
}

// Accept all network traffic to the coordinator
func (m *arangodb) AcceptCoordinatorTraffic() error {
	if api := m.nwBlocker; api == nil {
		return maskAny(fmt.Errorf("network-blocker not yet initialized"))
	} else {
		if err := api.AcceptTCP(m.coordinatorPort); err != nil {
			return maskAny(errgo.WithCausef(nil, err, "Failed to accept coordinator traffic"))
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
			return nil, maskAny(errgo.WithCausef(nil, err, "Failed to list network rules"))
		} else {
			return list, nil
		}
	}
}
