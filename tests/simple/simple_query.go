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
	created := false
	backoff := time.Millisecond * 250
	i := 0

	var err[] error
	var createResp[] util.ArangoResponse
	var cursorResp CursorResponse
	var createReqTime time.Time
	
	for {

		if time.Now().After(createTimeout) {
			break
		}
 	
		t.log.Infof("Creating (%d) AQL query cursor for '%s'...", i, c.name)
		queryReq := QueryRequest{
			//    Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN {d, s: SLEEP(10)}", collectionName),
			Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN d", c.name),
			BatchSize: 1,
			Count:     false,
		}
		
		createReqTime = time.Now()
		createResp, err := t.client.Post(
			"/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{0, 1, 201, 503},
			[]int{200, 202, 400, 404, 409, 307}, operationTimeout, 1)
		if err[0] != nil {
			// This is a failure
			break
		} else if createResp[0].StatusCode == 201 {
			created = true
			t.queryCreateCursorCounter.succeeded++
			t.log.Infof("Creating AQL cursor for collection '%s' succeeded", c.name)
			break
		}
		
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	if !created {
		t.queryCreateCursorCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create AQL cursor in collection '%s': %v", c.name, err[0]))
		return maskAny(err[0])
	}
			
	// Now continue fetching results.
	// This may fail if (and only if) the coordinator has changed.
	resultCount := len(cursorResp.Result)
	for {
		if !cursorResp.HasMore {
			// No more data
			break
		}

		// Wait a bit, so we increase the chance of a coordinator being restarting in between this actions.
		time.Sleep(time.Second * 5)

		// Fetch next results
		getResp, err := t.client.Put(
			"/_api/cursor/"+cursorResp.ID, nil, nil, nil, "", &cursorResp, []int{0, 1, 200, 404, 503},
			[]int{201, 202, 400, 409, 307}, operationTimeout, 1)
		if err[0] != nil {
			// This is a failure
			t.queryNextBatchCounter.failed++
			//t.reportFailure(test.NewFailure("Failed to read next AQL cursor batch in collection '%s': %v", c.name, err))
			t.log.Errorf("Failed to read next AQL cursor batch in collection '%s': %v", c.name, err[0])
			return maskAny(err[0])
		}

		// Check uptime of coordinator, if too short it has been rebooted since the initial query call.
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
				t.log.Infof("Reading next batch AQL cursor failed with 404, expected because of coordinator rebooted in between (%s)", createResp[0].CoordinatorURL)
				return nil
			}

			// Coordinator remains the same, this is a failure.
			t.queryNextBatchCounter.failed++
			t.reportFailure(test.NewFailure("Failed to read next AQL cursor batch in collection '%s' with same coordinator (%s): status 404", c.name, createResp[0].CoordinatorURL))
			return maskAny(fmt.Errorf("Status code 404"))
		}

		// Ok reading next batch succeeded
		t.queryNextBatchCounter.succeeded++
		resultCount += len(cursorResp.Result)
	}

	// We've fetched all documents, check result count
	if resultCount != 10 {
		t.reportFailure(test.NewFailure("Number of documents was %d, expected 10", resultCount))
		return maskAny(fmt.Errorf("Number of documents was %d, expected 10", resultCount))
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

	t.log.Infof("Creating long running AQL query for '%s'...", c.name)
	queryReq := QueryRequest{
		Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN {d:d, s:SLEEP(2)}", c.name),
		BatchSize: 10,
		Count:     false,
	}
	var cursorResp CursorResponse
	if _, err := t.client.Post("/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{201}, []int{200, 202, 400, 404, 409, 307}, operationTimeout, 1); err[0] != nil {
		// This is a failure
		t.queryLongRunningCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create long running AQL cursor in collection '%s': %v", c.name, err[0]))
		return maskAny(err[0])
	}
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
