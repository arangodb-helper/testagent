package simple

import (
	"fmt"
	"github.com/arangodb-helper/testagent/service/test"
	"time"
)

// createCollection creates a new collection.
// The operation is expected to succeed.
func (t *simpleTest) createView(name string, vtype string) error {

	view := struct {
		Name              string `json:"name"`
		Type              string `json:"type"`
	}{
		Name: name,
		Type: vtype,
	}

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 5)
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		t.log.Infof(
			"Creating (%d) view '%s' with numberOfShards=%d, replicationFactor=%d...", i, name, vtype)
		resp, err := t.client.Post(
			"/_api/view", nil, nil, view, "", nil, []int{0, 1, 200, 409, 500, 503},
			[]int{400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		// 0, 503: recheck without expectations
		//  exists:
		//    true: good (?) -> done
		//   false: retry
		//
		//      1:
		//  exists:
		//   true: failure
		//  false: retry
		//
		// 201:  good -> done
		//
		//  ...

		if err[0] != nil {
			// This is a failure
			t.createViewCounter.failed++
			t.reportFailure(test.NewFailure("Failed to create view '%s': %v", name, vtype, err[0]))
			return maskAny(err[0])
		} else {

			// do all verifications necessary

			
		}
		
		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.reportFailure(
		test.NewFailure("Timed out while trying to create (%d) view %s, %s.", i, name, vtype))
	return maskAny(fmt.Errorf("Timed out"))

	
}
