package simple

import (
	"fmt"
	"time"

	"github.com/arangodb/testAgent/service/test"
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

	operationTimeout, retryTimeout := time.Minute/3, time.Minute

	t.log.Infof("Creating AQL query cursor for '%s'...", c.name)
	queryReq := QueryRequest{
		//		Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN {d, s: SLEEP(10)}", collectionName),
		Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN d", c.name),
		BatchSize: 1,
		Count:     false,
	}
	var cursorResp CursorResponse
	createReqTime := time.Now()
	createResp, err := t.client.Post("/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{201}, []int{200, 202, 400, 404, 409, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.queryCreateCursorCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create AQL cursor in collection '%s': %v", c.name, err))
		return maskAny(err)
	}
	t.queryCreateCursorCounter.succeeded++
	t.log.Infof("Creating AQL cursor for collection '%s' succeeded", c.name)

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
		getResp, err := t.client.Put("/_api/cursor/"+cursorResp.ID, nil, nil, nil, "", &cursorResp, []int{200, 404}, []int{201, 202, 400, 409, 307}, operationTimeout, retryTimeout)
		if err != nil {
			// This is a failure
			t.queryNextBatchCounter.failed++
			t.reportFailure(test.NewFailure("Failed to read next AQL cursor batch in collection '%s': %v", c.name, err))
			return maskAny(err)
		}

		// Check uptime of coordinator, if too short it has been rebooted since the initial query call.
		uptime, err := t.getUptime(createResp.CoordinatorURL)
		if err != nil {
			t.log.Errorf("Failed to get uptime of server '%s': %v", createResp.CoordinatorURL, err)
		} else {
			t.log.Infof("Coordinator '%s' is up for %s", createResp.CoordinatorURL, uptime)
		}

		// Check status code
		if getResp.StatusCode == 404 {
			// Request failed, check if coordinator is different
			if createResp.CoordinatorURL != getResp.CoordinatorURL {
				// Coordinator changed, we expect this to fail now
				t.queryNextBatchNewCoordinatorCounter.succeeded++
				t.log.Infof("Reading next batch AQL cursor failed with 404, expected because of coordinator change (%s -> %s)", createResp.CoordinatorURL, getResp.CoordinatorURL)
				return nil
			}

			// Check uptime
			if uptime < time.Since(createReqTime) && uptime != 0 {
				// Coordinator rebooted, we expect this to fail now
				t.queryNextBatchNewCoordinatorCounter.succeeded++
				t.log.Infof("Reading next batch AQL cursor failed with 404, expected because of coordinator rebooted in between (%s)", createResp.CoordinatorURL)
				return nil
			}

			// Coordinator remains the same, this is a failure.
			t.queryNextBatchCounter.failed++
			t.reportFailure(test.NewFailure("Failed to read next AQL cursor batch in collection '%s' with same coordinator (%s): status 404", c.name, createResp.CoordinatorURL))
			return maskAny(fmt.Errorf("Status code 404"))
		} else if getResp.StatusCode == 200 {
			// Request succeeded, check if coordinator is same as create-cursor request
			if createResp.CoordinatorURL != getResp.CoordinatorURL {
				// Coordinator changed, we expected a failure, but got a success. Not good.
				t.queryNextBatchNewCoordinatorCounter.failed++
				t.reportFailure(test.NewFailure("Reading next batch AQL cursor succeeded with 200, but expected a 404 because of coordinator change"))
				t.log.Infof("Reading next batch AQL cursor succeeded with 200, but expected a 404 because of coordinator change")
				return nil
			}
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

	operationTimeout, retryTimeout := time.Minute/2, time.Minute*2

	t.log.Infof("Creating long running AQL query for '%s'...", c.name)
	queryReq := QueryRequest{
		Query:     fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN {d:d, s:SLEEP(2)}", c.name),
		BatchSize: 10,
		Count:     false,
	}
	var cursorResp CursorResponse
	if _, err := t.client.Post("/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{201}, []int{200, 202, 400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.queryLongRunningCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create long running AQL cursor in collection '%s': %v", c.name, err))
		return maskAny(err)
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
	operationTimeout, retryTimeout := time.Minute/2, time.Minute*2
	var statsResp struct {
		Server struct {
			Uptime float64 `json:"uptime"`
		} `json:"server"`
	}
	if err := t.client.SetCoordinator(coordinatorURL); err != nil {
		t.log.Errorf("Failed to set coordinator URL to '%s': %v", coordinatorURL, err)
		return 0, maskAny(err)
	}
	if resp, err := t.client.Get("/_admin/statistics", nil, nil, &statsResp, []int{200}, []int{201, 202, 400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		return 0, maskAny(fmt.Errorf("Failed to query uptime of '%s': %v", resp.CoordinatorURL, err))
	} else if resp.CoordinatorURL != coordinatorURL {
		return 0, maskAny(fmt.Errorf("Failed to query uptime of '%s': got response from other coordinator '%s'", coordinatorURL, resp.CoordinatorURL))
	}

	return time.Second * time.Duration(statsResp.Server.Uptime), nil
}
