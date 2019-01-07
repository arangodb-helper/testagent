package arangodb

import "sync"

type portSpace struct {
	mutex     sync.Mutex
	basePort  int
	portDelta int
	inUse     map[string]int
}

// Initialize the structure.
// Call only once.
func (s *portSpace) Initialize(basePort, portDelta int) {
	s.basePort = basePort
	s.portDelta = portDelta
	s.inUse = make(map[string]int)
}

// Allocate a port for given ID
func (s *portSpace) Allocate(id string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	port := s.basePort + s.portDelta
	for {
		found := false
		for _, p := range s.inUse {
			if p == port {
				found = true
				break
			}
		}
		if !found {
			s.inUse[id] = port
			return port
		}
		port += s.portDelta
	}
}

// Release the port allocated for the given ID
func (s *portSpace) Release(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.inUse, id)
}
