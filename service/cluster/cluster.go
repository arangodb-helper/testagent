package cluster

type ClusterBuilder interface {
	// Create creates and starts a new cluster.
	// The number of "machines" created equals the given agency size.
	// This function returns when the cluster is operational (or an error occurs)
	Create(agencySize int) (Cluster, error)
}

type Cluster interface {
	// RestartMachine restarts a single machine in the cluster
	RestartMachine() error

	// Remove the entire cluster
	Destroy() error
}
