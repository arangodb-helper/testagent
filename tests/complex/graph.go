package complex

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

type EdgeDocument struct {
	TestDocument
	Value   int64  `json:"value"`
	Payload string `json:"payload"`
	From    string `json:"_from"`
	To      string `json:"_to"`
}

func NewEdgeDocument(from string, to string, vertexColName string, edgeSize int, seed int64) EdgeDocument {
	randGen := rand.New(rand.NewSource(seed))
	payloadBytes := make([]byte, edgeSize)
	lowerBound := 32
	upperBound := 126
	for i := 0; i < edgeSize; i++ {
		payloadBytes[i] = byte(randGen.Int31n(int32(upperBound-lowerBound)) + int32(lowerBound))
	}
	return EdgeDocument{
		TestDocument: TestDocument{
			Seed:          seed,
			UpdateCounter: 0,
		},
		Value:   seed,
		Payload: string(payloadBytes),
		From:    vertexColName + "/" + from,
		To:      vertexColName + "/" + to,
	}
}

// createGraph creates a new graph.
// The operation is expected to succeed.
func (t *ComplextTest) createGraph(graphName string,
	edgeCol string, fromCols []string, toCols []string, orphans []string,
	isSmart bool, isDisjoint bool, smartGraphAttribute string,
	satellites []string, numberOfShards int, replicationFactor int, writeConcern int) error {
	opts := struct {
		Name            string `json:"name"`
		EdgeDefinitions []struct {
			Collection string   `json:"collection"`
			From       []string `json:"from"`
			To         []string `json:"to"`
		} `json:"edgeDefinitions"`
		OrphanCollections []string `json:"orphanCollections,omitempty"`
		Options           struct {
			IsSmart             bool     `json:"isSmart"`
			IsDisjoint          bool     `json:"isDisjoint"`
			SmartGraphAttribute string   `json:"smartGraphAttribute,omitempty"`
			Satellites          []string `json:"satellites,omitempty"`
			NumberOfShards      int      `json:"numberOfShards,omitempty"`
			ReplicationFactor   int      `json:"replicationFactor,omitempty"`
			WriteConcern        int      `json:"writeConcern,omitempty"`
		} `json:"options"`
	}{
		Name: graphName,
		EdgeDefinitions: []struct {
			Collection string   `json:"collection"`
			From       []string `json:"from"`
			To         []string `json:"to"`
		}{{
			Collection: edgeCol,
			From:       fromCols,
			To:         toCols,
		}},
		OrphanCollections: orphans,
		Options: struct {
			IsSmart             bool     `json:"isSmart"`
			IsDisjoint          bool     `json:"isDisjoint"`
			SmartGraphAttribute string   `json:"smartGraphAttribute,omitempty"`
			Satellites          []string `json:"satellites,omitempty"`
			NumberOfShards      int      `json:"numberOfShards,omitempty"`
			ReplicationFactor   int      `json:"replicationFactor,omitempty"`
			WriteConcern        int      `json:"writeConcern,omitempty"`
		}{
			IsSmart:             isSmart,
			IsDisjoint:          isDisjoint,
			SmartGraphAttribute: smartGraphAttribute,
			Satellites:          satellites,
			NumberOfShards:      numberOfShards,
			ReplicationFactor:   replicationFactor,
			WriteConcern:        writeConcern,
		},
	}
	//operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	// For now, we increase the timeout to 5 minutes, since the cluster-internal
	// timeout is 4 minutes:
	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 5)

	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		checkRetry := false
		success := false
		shouldNotExist := false
		shouldExist := false

		t.log.Infof("Creating a graph named %s...", graphName)
		q := url.Values{}
		q.Set("waitForSync", "true")
		url := "/_api/gharial"
		resp, err := t.client.Post(
			url, q, nil, opts, "", nil, []int{0, 1, 201, 202, 409, 500, 503},
			[]int{400, 403, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		// 0, 503: recheck without expectations
		//     there: good
		//     not there: retry
		// 20x   : good
		// 1, 500: graph creation couldn't be finished.
		//     there: failure
		//     not there: retry
		// 409   :
		//     first attempt: failure
		//     later attempts:
		//     recheck
		//         there: done
		//         else : failure

		if err[0] == nil {
			if resp[0].StatusCode == 201 || resp[0].StatusCode == 202 {
				success = true
			} else {
				checkRetry = true
				if resp[0].StatusCode == 1 || resp[0].StatusCode == 500 { // connection refused or not created
					shouldNotExist = true
				} else if resp[0].StatusCode == 409 {
					if i == 1 {
						// This is a failure
						t.createGraphCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(), "Failed to create graph '%s': got 409 on first attempt", graphName))
						return maskAny(fmt.Errorf("Failed to create graph '%s': got 409 on first attempt", graphName))
					} else {
						shouldExist = true
					}
				}
			}
		} else {
			// This is a failure
			t.createGraphCounter.failed++
			t.reportFailure(test.NewFailure(t.Name(), "Failed to create graph '%s': %v", graphName, err[0]))
			return maskAny(err[0])
		}

		if checkRetry {

			t.log.Infof("Checking existence of graph '%s' ...", graphName)
			exists, checkErr := t.graphExists(graphName)
			t.log.Infof("... got result %v and error %v", exists, checkErr)

			if checkErr == nil {
				if exists {
					if shouldNotExist {
						// This is a failure
						t.createGraphCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(),
							"Failed to create graph '%s' rechecked and failed existence", graphName))
						return maskAny(fmt.Errorf("Failed to create  '%s' rechecked and failed existence", graphName))
					}
					success = true
				} else {
					if shouldExist {
						// This is a failure
						t.createGraphCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(),
							"Failed to create graph '%s' rechecked and failed existence", graphName))
						return maskAny(fmt.Errorf("Failed to create graph '%s' rechecked and failed existence", graphName))
					}
				}
			} else {
				return maskAny(checkErr)
			}
		}

		if success {
			t.createGraphCounter.succeeded++
			t.log.Infof(
				"Creating graph '%s' succeeded", graphName)
			return nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.reportFailure(
		test.NewFailure(t.Name(), "Timed out while trying to create (%d) graph %s.", i, graphName))
	return maskAny(fmt.Errorf("Timed out while trying to create (%d) graph %s.", i, graphName))
}

func (t *ComplextTest) graphExists(graphName string) (bool, error) {

	operationTimeout := time.Duration(ReadTimeout) * time.Second
	timeout := time.Now().Add(operationTimeout)

	i := 0
	backoff := time.Millisecond * 250
	url := fmt.Sprintf("/_api/gharial/%s", graphName)

	for {

		i++
		if time.Now().After(timeout) {
			break
		}

		t.log.Infof("Checking (%d) graph '%s'...", i, graphName)
		resp, err := t.client.Get(
			url, nil, nil, nil, []int{0, 1, 200, 404, 503}, []int{400, 409, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d", resp[0].StatusCode, resp[0].Error_.ErrorNum)

		if err[0] != nil {
			// This is a failure
			t.log.Infof("Failed checking for graph '%s': %v", graphName, err[0])
			return false, maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			return false, nil
		} else if resp[0].StatusCode == 200 {
			return true, nil
		}

		// 0, 1, 503 retry
		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// This is a failure
	out := fmt.Errorf("Timed out checking for graph '%s'", graphName)
	t.log.Error(out)
	return false, maskAny(out)

}

func (t *ComplextTest) dropGraph(graphName string, dropCollections bool) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(t.OperationTimeout * 5)

	backoff := time.Millisecond * 250
	i := 0

	success := false
	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		q := url.Values{}
		url := fmt.Sprintf("/_api/gharial/%s", graphName)
		q.Set("dropCollections", strconv.FormatBool(dropCollections))

		t.log.Infof("Removing (%d) graph '%s'...", i, graphName)
		resp, err := t.client.Delete(
			url, q, nil, []int{0, 1, 201, 202, 404, 500, 503}, []int{400, 409, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d", resp[0].StatusCode, resp[0].Error_.ErrorNum)

		if err[0] != nil {
			// This is a failure
			t.dropGraphCounter.failed++
			t.reportFailure(test.NewFailure(t.Name(), "Failed to drop graph '%s': %v", graphName, err[0]))
			return maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			// graph not found.
			// This can happen if the first attempt timed out, but did actually succeed.
			// So we accept this if there are multiple attempts.
			if i == 1 { // this is a failure in first run
				// Not enough attempts, this is a failure
				t.dropGraphCounter.failed++
				t.reportFailure(
					test.NewFailure(t.Name(), "Failed to drop graph '%s': got 404 after only 1 attempt", graphName))
				return maskAny(fmt.Errorf("Failed to drop graph '%s': got 404 after only 1 attempt", graphName))
			} else {
				success = true
			}
		} else if resp[0].StatusCode == 200 {
			success = true
		}

		if success {
			t.dropGraphCounter.succeeded++
			t.log.Infof("Droping graph '%s' succeeded", graphName)
			return nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	t.dropGraphCounter.failed++
	t.reportFailure(test.NewFailure(t.Name(), "Timed out (%d) while droping graph '%s'", i, graphName))
	return maskAny(fmt.Errorf("Timed out (%d) while droping graph '%s'", i, graphName))

}

func (t *GraphTest) createEdge(to string, from string, edgeColName string, vertexColName string, edgeSize int) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout)

	var seed = t.documentIdSeq
	t.documentIdSeq++
	document := NewEdgeDocument(from, to, vertexColName, edgeSize, seed)

	q := url.Values{}
	q.Set("waitForSync", "true")
	url := fmt.Sprintf("/_api/document/%s", edgeColName)
	backoff := time.Millisecond * 250
	i := 0
	mustNotExist := true

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		checkRetry := false
		success := false

		t.log.Infof("Creating edge from '%s' to '%s' in collection '%s'.", from, to, edgeColName)
		resp, err := t.client.Post(url, q, nil, document, "", nil,
			[]int{0, 1, 200, 201, 202, 404, 409, 410, 503}, []int{400, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] == nil { // we have a response
			if resp[0].StatusCode == 503 || resp[0].StatusCode == 409 || resp[0].StatusCode == 0 {
				// 0, 503 and 409 -> check if accidentally successful
				checkRetry = true
				mustNotExist = false
			} else if resp[0].StatusCode == 202 {
				// should not respond 202 if waitForSync == true
				t.edgeDocumentCreateCounter.failed++
				errorMsg := fmt.Sprintf("Server responded with status code 202 to a request with waitForSync=true: %v", resp[0])
				t.reportFailure(
					test.NewFailure(t.Name(), errorMsg))
				return maskAny(fmt.Errorf(errorMsg))
			} else if resp[0].StatusCode == 410 {
				// retry. if 410 was returned on first attempt, document must not exist
				checkRetry = true
			} else if resp[0].StatusCode == 404 {
				// 404: If transaction was lost(error 1655) due to server restart, then we should just retry.
				// In any other case(e.g. collection not found etc. - fail.)
				if resp[0].Error_.ErrorNum != 1655 {
					t.edgeDocumentCreateCounter.failed++
					t.reportFailure(
						test.NewFailure(t.Name(), "Failed to create edge in collection '%s': unexpected response %v", edgeColName, resp[0]))
					return maskAny(fmt.Errorf("Failed to create edge in collection '%s': unexpected response %v", edgeColName, resp[0]))
				}
			} else if resp[0].StatusCode != 1 {
				document.Rev = resp[0].Rev
				success = true
				mustNotExist = false
			}
		} else { // failure
			t.edgeDocumentCreateCounter.failed++
			t.reportFailure(
				test.NewFailure(t.Name(), "Failed to create edge in collection '%s': %v", edgeColName, err[0]))
			return maskAny(err[0])
		}

		if checkRetry {
			edge, err := readDocumentBySeed(&t.ComplextTest, edgeColName, seed)
			if err == nil && edge != nil {
				if !mustNotExist {
					document.TestDocument = *edge
					success = true
				} else {
					t.edgeDocumentCreateCounter.failed++
					t.reportFailure(
						test.NewFailure(t.Name(),
							"Incorrect behaviour during edge creation in collection '%s'. "+
								"Was not expecting the document to be created, but it was created.",
							edgeColName))
					return errors.New("Incorrect behaviour during edge creation. Was not expecting the document to be created, but it was created.")

				}
			} else {
				success = false
			}
		}

		if success {
			t.existingEdgeDocuments = append(t.existingEdgeDocuments, document.TestDocument)
			t.edgeDocumentCreateCounter.succeeded++
			t.log.Infof("Creating document in '%s' succeeded", edgeColName)
			return nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.edgeDocumentCreateCounter.failed++
	t.reportFailure(
		test.NewFailure(t.Name(), "Timed out while trying to create a document in '%s'.", edgeColName))
	return maskAny(fmt.Errorf("Timed out while trying to create a document in '%s'.", edgeColName))
}

func lengthExcludingNils(arr []any) int {
	length := 0
	for i := 0; i < len(arr); i++ {
		if arr[i] != nil {
			length++
		}
	}
	return length
}

func (t *GraphTest) traverse(to string, from string, graphName string, expectedLength int) error {
	operationTimeout := t.OperationTimeout * 4
	testTimeout := time.Now().Add(time.Minute * 15)

	i := 0
	backoff := time.Millisecond * 100

	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++

		t.log.Infof("Creating (%d) long running AQL query to traverse graph '%s'...", i, graphName)
		query := fmt.Sprintf(`FOR v, e IN OUTBOUND SHORTEST_PATH "%s" TO "%s" GRAPH "%s" RETURN e`, from, to, graphName)
		t.log.Debugf("Running query: %s", query)
		queryReq := QueryRequest{
			Query:     query,
			BatchSize: expectedLength * 2,
			Count:     false,
		}
		var cursorResp CursorResponse
		resp, err := t.client.Post(
			"/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{0, 1, 201, 410, 500, 503},
			[]int{200, 202, 400, 404, 409, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] != nil {
			// This is a failure
			t.traverseGraphCounter.failed++
			t.reportFailure(test.NewFailure(t.Name(),
				"Failed to traverse graph '%s': %v", graphName, err[0]))
			return maskAny(err[0])
		}

		if resp[0].StatusCode == 201 { // 0, 1, 503 just go another round
			actualLength := lengthExcludingNils(cursorResp.Result)
			hasMore := cursorResp.HasMore

			t.log.Infof("Creating long running AQL query for collection '%s' succeeded", graphName)
			// We should've fetched all documents, check result count
			if !(actualLength == expectedLength && !hasMore) {
				t.reportFailure(test.NewFailure(t.Name(), "Graph traversal failed: was expecting a chain of %d edges, got %d. Query: %s", expectedLength, actualLength, query))
				t.traverseGraphCounter.failed++
				return maskAny(fmt.Errorf("Graph traversal failed: was expecting a chain of %d edges, got %d. Query: %s", expectedLength, actualLength, query))
			}
			t.traverseGraphCounter.succeeded++
			return nil
		}

		// Otherwise we fall through and simply try again.

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.traverseGraphCounter.failed++
	t.reportFailure(
		test.NewFailure(t.Name(), "Timed out while trying to traverse from '%s' to '%s' in graph '%s'.", from, to, graphName))
	return maskAny(fmt.Errorf("Timed out while trying to traverse from '%s' to '%s' in graph '%s'.", from, to, graphName))
}
