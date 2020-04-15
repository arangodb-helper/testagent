package chaos

import "github.com/arangodb-helper/testagent/service/cluster"

// MachineList is a list of cluster machines
type MachineList []cluster.Machine

// Contains returns true if the given machine is found (by its ID) in this list, false otherwise.
func (l MachineList) Contains(m cluster.Machine) bool {
	for _, x := range l {
		if x.ID() == m.ID() {
			return true
		}
	}
	return false
}

// Except returns a new machine list with all entries from this list
// that are not in the given toRemove list.
func (l MachineList) Except(toRemove MachineList) MachineList {
	var result MachineList
	for _, m := range l {
		if !toRemove.Contains(m) {
			result = append(result, m)
		}
	}
	return result
}

// Union returns a new machine list with all entries from this list
// and the other list combined.
func (l MachineList) Union(other MachineList) MachineList {
	var result MachineList
	result = append(result, l...)
	for _, m := range other {
		if !l.Contains(m) {
			result = append(result, m)
		}
	}
	return result
}

// Intersection returns a new machine list with all entries that are contained in both lists.
func (l MachineList) Intersection(other MachineList) MachineList {
	var result MachineList
	for _, m := range l {
		if other.Contains(m) {
			result = append(result, m)
		}
	}
	return result
}

// ExceptAgents returns a new machine list with all entries from this list
// that do not have an agent.
func (l MachineList) ExceptAgents() MachineList {
	var result MachineList
	for _, m := range l {
		if !m.HasAgent() {
			result = append(result, m)
		}
	}
	return result
}
