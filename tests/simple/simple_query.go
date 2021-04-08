package simple

import (
	"fmt"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
)

type QueryRequest struct {
	Query     string `json:"query"`
	BatchSize int    `json:"batchSize,omitempty"`
	Count     bool   `json:"count,omitempty"`
}

type CursorResponse struct {
	HasMore bool          `json:"hasMore,omitempty"`
	ID      string        `json:"id,omitempty"`
	Result  []interface{} `json:"result,omitempty"`
}

// queryDocuments runs an AQL query.
// The operation is expected to succeed.
func (t *simpleTest) queryDocuments(c *collection) error {
	if len(c.existingDocs) < 10 {
		t.log.Infof("Skipping query test, we need 10 or more documents")
		return nil
	}

	operationTimeout := t.OperationTimeout
	createTimeout := time.Now().Add(operationTimeout * 5)
	backoff := time.Millisecond * 250
	i := 0

	var err []error
	var createResp []util.ArangoResponse
	var cursorResp CursorResponse
	var createReqTime time.Time

	for {

		if time.Now().After(createTimeout) {
			break
		}
		i++

		t.log.Infof("Creating (%d) AQL query cursor for '%s'...", i, c.name)
		queryReq := QueryRequest{
			Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN d", c.name),
			BatchSize: 1,
			Count:     false,
		}

		createReqTime = time.Now()
		createResp, err = t.client.Post(
			"/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{0, 1, 201, 500, 503},
			[]int{200, 202, 307, 400, 404, 409}, operationTimeout, 1)
		if err[0] != nil {
			// This is a failure
			t.queryCreateCursorCounter.failed++
			t.reportFailure(test.NewFailure("Failed to create AQL cursor in collection '%s': %v", c.name, err[0]))
			return maskAny(err[0])
		} else if createResp[0].StatusCode == 201 {
			t.queryCreateCursorCounter.succeeded++
			t.log.Infof("Creating AQL cursor for collection '%s' succeeded", c.name)
			break
		}

		// Otherwise we fall through and simply try again. Note that if an
		// attempt times out it is OK to simply retry, even if the old one
		// eventually gets through, we can then simply work with the new
		// cursor. Furthermore note that we found that currently the undocumented
		// error code 500 can happen if a dbserver suffers from some chaos
		// during cursor creation. We can simply retry, too.

		t.log.Infof("Creating (%d) AQL query cursor for '%s' got %d", i, c.name, createResp[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	// Now continue fetching results.
	// This may fail if (and only if) the coordinator has changed.
	resultCount := len(cursorResp.Result)
	nrTimeOuts := 0
	for {
		if !cursorResp.HasMore {
			// No more data, now see if we have the right amount of results
			break
		}

    // Wait a bit, so we increase the chance of a coordinator being
    // restarting in between this actions (or some other chaos to happen).
		time.Sleep(time.Second * 5)

		// Fetch next results
		getResp, err := t.client.Put(
			"/_api/cursor/"+cursorResp.ID, nil, nil, nil, "", &cursorResp, []int{0, 1, 200, 404, 500, 503},
			[]int{201, 202, 400, 409, 307}, operationTimeout, 1)
		if err[0] != nil {
			// This is a failure
			t.queryNextBatchCounter.failed++
			t.reportFailure(test.NewFailure(
				"Failed to read next AQL cursor batch in collection '%s': %v", c.name, err[0]))
			t.log.Errorf(
				"Failed to read next AQL cursor batch in collection '%s': %v", c.name, err[0])
			return maskAny(err[0])
		}

		// This loop is a loop to fetch all 10 results and at the same time
		// a retry loop. In the cases 0, 1 and 503 we must simply try again.
		// Note however, that a timeout (0) or 503 can mean that a result
		// was consumed but not counted by us. Therefore, we count these
		// occurrences and are more tolerant at the end w.r.t. the final
		// count.

    // Check uptime of coordinator, if too short it has been rebooted
    // since the initial query call.
		uptime, oerr := t.getUptime(createResp[0].CoordinatorURL)
		if err != nil {
			t.log.Errorf("Failed to get uptime of server '%s': %v", createResp[0].CoordinatorURL, oerr)
		} else {
			t.log.Infof("Coordinator '%s' is up for %s", createResp[0].CoordinatorURL, uptime)
		}

		// Check status code
		if getResp[0].StatusCode == 404 {

			// Check uptime
			if uptime < time.Since(createReqTime) {
				// Note that if the uptime call failed, we can get 0 for uptime, which is OK,
				// since this probably means that the coordinator is not yet up again.
				// Coordinator rebooted, we expect this to fail now
				t.queryNextBatchNewCoordinatorCounter.succeeded++
				t.log.Infof(
					"Reading next batch AQL cursor failed with 404, expected because of rebooted coord in between (%s)",
					createResp[0].CoordinatorURL)
				return nil
			}

			// Coordinator remains the same, this is a failure.
			t.queryNextBatchCounter.failed++
			t.reportFailure(test.NewFailure("Failed to read next AQL cursor batch in collection '%s' with same coordinator (%s): status 404", c.name, createResp[0].CoordinatorURL))
			return maskAny(fmt.Errorf("Status code 404"))
		} else if getResp[0].StatusCode == 500 {
			// A dbserver might have suffered from chaos, and then the query
			// is lost. We simply give up in this case. Document count must
			// not be considered then. Furthermore, we have no idea which
			// dbserver could have been responsible, so we cannot check
			// its uptime in a sensible way.
			t.queryNextBatchCounter.failed++
			t.log.Infof(
					"Reading next batch AQL cursor failed with 500, expected because of potential chaos with dbservers in between (coordinator: %s)",
				createResp[0].CoordinatorURL)
			return nil
		} else if getResp[0].StatusCode == 0 || getResp[0].StatusCode == 503 {
			nrTimeOuts++
		}

		// Ok reading next batch succeeded
		t.queryNextBatchCounter.succeeded++
		resultCount += len(cursorResp.Result)
	}

	// We've fetched all documents, check result count
	if resultCount > 10 || resultCount < 10 - nrTimeOuts {
		t.reportFailure(test.NewFailure("Number of documents was %d, expected 10", resultCount))
		return maskAny(fmt.Errorf("Number of documents was %d, expected 10", resultCount))
	}

	if resultCount != 10 {
		t.log.Infof("Got a different result than 10: %d, but this is explained by timeouts we got: %d.", resultCount, nrTimeOuts)
	}
	return nil
}

// queryDocumentsLongRunning runs a long running AQL query.
// The operation is expected to succeed.
func (t *simpleTest) queryDocumentsLongRunning(c *collection) error {
	if len(c.existingDocs) < 10 {
		t.log.Infof("Skipping query test, we need 10 or more documents")
		return nil
	}

	operationTimeout := t.OperationTimeout*2
	testTimeout := time.Now().Add(operationTimeout * 4)
	i := 0
	backoff := time.Millisecond * 100

	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++

		t.log.Infof("Creating (%d) long running AQL query for '%s'...", i, c.name)
		queryReq := QueryRequest{
			Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN {d:d, s:SLEEP(2)}", c.name),
			BatchSize: 10,
			Count:     false,
		}
		var cursorResp CursorResponse
		resp, err := t.client.Post(
			"/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{0, 1, 201, 500, 503},
			[]int{200, 202, 400, 404, 409, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.queryLongRunningCounter.failed++
			t.reportFailure(test.NewFailure(
				"Failed to create long running AQL cursor in collection '%s': %v", c.name, err[0]))
			return maskAny(err[0])
		}

		if resp[0].StatusCode == 201 { // 0, 1, 503 just go another round
			resultCount := len(cursorResp.Result)
			t.queryLongRunningCounter.succeeded++
			t.log.Infof("Creating long running AQL query for collection '%s' succeeded", c.name)
			// We should've fetched all documents, check result count
			if resultCount != 10 {
				t.reportFailure(test.NewFailure("Number of documents was %d, expected 10", resultCount))
				return maskAny(fmt.Errorf("Number of documents was %d, expected 10", resultCount))
			}
			return nil
		}

		// Otherwise we fall through and simply try again.

		t.log.Infof("Creating (%d) long running AQL query for '%s' got %d", i, c.name, resp[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	t.reportFailure(test.NewFailure(
		"Timed out create (%d) long running AQL cursor in collection '%s'", i, c.name))
	return maskAny(fmt.Errorf(
		"Timed out create (%d) long running AQL cursor in collection '%s'", i, c.name))

}

// getUptime queries the uptime of the given coordinator.
func (t *simpleTest) getUptime(coordinatorURL string) (time.Duration, error) {
	t.log.Infof("Checking uptime of '%s'", coordinatorURL)
	operationTimeout := t.OperationTimeout*2
	var statsResp struct {
		Server struct {
			Uptime float64 `json:"uptime"`
		} `json:"server"`
	}
	if err := t.client.SetCoordinator(coordinatorURL); err != nil {
		t.log.Errorf("Failed to set coordinator URL to '%s': %v", coordinatorURL, err)
		return 0, maskAny(err)
	}
	if resp, err := t.client.Get("/_admin/statistics", nil, nil, &statsResp, []int{200}, []int{201, 202, 400, 404, 409, 307}, operationTimeout, 1); err[0] != nil {
		return 0, maskAny(fmt.Errorf("Failed to query uptime of '%s': %v", resp[0].CoordinatorURL, err[0]))
	} else if resp[0].CoordinatorURL != coordinatorURL {
		return 0, maskAny(fmt.Errorf("Failed to query uptime of '%s': got response from other coordinator '%s'", coordinatorURL, resp[0].CoordinatorURL))
	}

	return time.Second * time.Duration(statsResp.Server.Uptime), nil
}
