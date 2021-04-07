package simple

import (
	"fmt"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

// queryUpdateDocuments runs an AQL update query.
// The operation is expected to succeed.
func (t *simpleTest) queryUpdateDocuments(c *collection, key string) (string, error) {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 5)
	backoff := time.Millisecond * 250
	i := 0
	
	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++
		
		t.log.Infof("Creating update AQL query for collection '%s'...", c.name)
		newName := fmt.Sprintf("AQLUpdate name %s", time.Now())
		queryReq := QueryRequest{
			Query:     fmt.Sprintf("UPDATE \"%s\" WITH { name: \"%s\" } IN %s RETURN NEW", key, newName, c.name),
			BatchSize: 1,
			Count:     false,
		}
		
		var cursorResp CursorResponse
		resultDocument := &UserDocument{}
		cursorResp.Result = []interface{}{resultDocument}
		resp, err := t.client.Post(
			"/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{0, 1, 201, 409, 503},
			[]int{200, 202, 400, 404, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.queryUpdateCounter.failed++
			t.reportFailure(test.NewFailure(
				"Failed to create update AQL cursor in collection '%s': %v", c.name, err[0]))
			return "", maskAny(err[0])
		}

		if resp[0].StatusCode == 201 {
			resultCount := len(cursorResp.Result)
			if resultCount != 1 {
				// This is a failure
				t.queryUpdateCounter.failed++
				t.reportFailure(test.NewFailure(
					"Failed to create update AQL cursor in collection '%s': expected 1 result, got %d", c.name, resultCount))
				return "", maskAny(fmt.Errorf(
					"Number of documents was %d, expected 1", resultCount))
			}

			// Update document
			c.existingDocs[key] = *resultDocument
			t.queryUpdateCounter.succeeded++
			t.log.Infof("Creating update AQL query for collection '%s' succeeded", c.name)

			return resultDocument.rev, nil
		}
			
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}
	}

	t.queryUpdateCounter.failed++
	t.reportFailure(test.NewFailure(
		"Timed out while creating (%d) update AQL cursor in collection '%s'",	i, c.name))
	return "", maskAny(fmt.Errorf(
		"Timed out while creating (%d) update AQL cursor in collection '%s'",	i, c.name))

}

// queryUpdateDocumentsLongRunning runs a long running AQL update query.
// The operation is expected to succeed.
func (t *simpleTest) queryUpdateDocumentsLongRunning(c *collection, key string) (string, error) {
	operationTimeout := t.OperationTimeout*3

	t.log.Infof("Creating long running update AQL query for collection '%s'...", c.name)
	newName := fmt.Sprintf("AQLLongRunningUpdate name %s", time.Now())
	queryReq := QueryRequest{
		Query:     fmt.Sprintf("UPDATE \"%s\" WITH { name: \"%s\", unknown: SLEEP(15) } IN %s RETURN NEW", key, newName, c.name),
		BatchSize: 1,
		Count:     false,
	}
	var cursorResp CursorResponse
	resultDocument := &UserDocument{}
	cursorResp.Result = []interface{}{resultDocument}
	if _, err := t.client.Post("/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{201, 409}, []int{200, 202, 400, 404, 307}, operationTimeout, 1); err[0] != nil {
		// This is a failure
		t.queryUpdateLongRunningCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create long running update AQL cursor in collection '%s': %v", c.name, err[0]))
		return "", maskAny(err[0])
	}
	resultCount := len(cursorResp.Result)
	if resultCount != 1 {
		// This is a failure
		t.queryUpdateLongRunningCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create long running update AQL cursor in collection '%s': expected 1 result, got %d", c.name, resultCount))
		return "", maskAny(fmt.Errorf("Number of documents was %d, expected 1", resultCount))
	}

	// Update document
	c.existingDocs[key] = *resultDocument
	t.queryUpdateLongRunningCounter.succeeded++
	t.log.Infof("Creating long running update AQL query for collection '%s' succeeded", c.name)

	return resultDocument.rev, nil
}
