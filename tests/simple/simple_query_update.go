package simple

import (
	"fmt"
	"time"

	"github.com/arangodb/testAgent/service/test"
)

// queryUpdateDocuments runs an AQL update query.
// The operation is expected to succeed.
func (t *simpleTest) queryUpdateDocuments(collectionName, key string) (string, error) {
	if len(t.existingDocs) < 10 {
		t.log.Infof("Skipping query test, we need 10 or more documents")
		return "", nil
	}

	operationTimeout, retryTimeout := time.Minute/3, time.Minute

	t.log.Infof("Creating update AQL query for collection '%s'...", collectionName)
	newName := fmt.Sprintf("AQLUpdate name %s", time.Now())
	queryReq := QueryRequest{
		Query:     fmt.Sprintf("UPDATE \"%s\" WITH { name: \"%s\" } IN %s RETURN NEW", key, newName, collectionName),
		BatchSize: 1,
		Count:     false,
	}
	var cursorResp CursorResponse
	resultDocument := &UserDocument{}
	cursorResp.Result = []interface{}{resultDocument}
	if _, err := t.client.Post("/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{201}, []int{200, 202, 400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.queryUpdateCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create update AQL cursor in collection '%s': %v", collectionName, err))
		return "", maskAny(err)
	}
	resultCount := len(cursorResp.Result)
	if resultCount != 1 {
		// This is a failure
		t.queryUpdateCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create update AQL cursor in collection '%s': expected 1 result, got %d", collectionName, resultCount))
		return "", maskAny(fmt.Errorf("Number of documents was %d, expected 1", resultCount))
	}

	// Update document
	t.existingDocs[key] = *resultDocument
	t.queryUpdateCounter.succeeded++
	t.log.Infof("Creating update AQL query for collection '%s' succeeded", collectionName)

	return resultDocument.rev, nil
}
