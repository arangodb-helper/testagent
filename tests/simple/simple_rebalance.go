package simple

import "github.com/arangodb/testAgent/service/test"

// rebalanceShards attempts to rebalance shards over the existing servers.
// The operation is expected to succeed.
func (t *simpleTest) rebalanceShards() error {
	opts := struct{}{}
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	t.log.Infof("Rebalancing shards...")
	if _, err := t.client.Post("/_admin/cluster/rebalanceShards", nil, nil, opts, "", nil, []int{202}, []int{400, 403, 503}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.rebalanceShardsCounter.failed++
		t.reportFailure(test.NewFailure("Failed to rebalance shards: %v", err))
		return maskAny(err)
	}
	t.rebalanceShardsCounter.succeeded++
	t.log.Infof("Rebalancing shards succeeded")
	return nil
}
