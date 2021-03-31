package simple

import "github.com/arangodb-helper/testagent/service/test"

// rebalanceShards attempts to rebalance shards over the existing servers.
// The operation is expected to succeed.
func (t *simpleTest) rebalanceShards() error {
	opts := struct{}{}
	operationTimeout := t.OperationTimeout
	t.log.Infof("Rebalancing shards...")
	if _, err := t.client.Post("/_admin/cluster/rebalanceShards", nil, nil, opts, "", nil, []int{202}, []int{400, 403, 503}, operationTimeout, 1); err[0] != nil {
		// This is a failure
		t.rebalanceShardsCounter.failed++
		t.reportFailure(test.NewFailure("Failed to rebalance shards: %v", err))
		return maskAny(err[0])
	}
	t.rebalanceShardsCounter.succeeded++
	t.log.Infof("Rebalancing shards succeeded")
	return nil
}
