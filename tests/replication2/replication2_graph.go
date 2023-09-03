package replication2

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

type EdgeDocument struct {
	Rev           string `json:"_rev,omitempty"`
	Value         int64  `json:"value"`
	UpdateCounter int    `json:"update_counter"`
	Payload       string `json:"payload"`
	From          string `json:"_from"`
	To            string `json:"_to"`
}

func NewEdgeDocument(from string, to string, colName string) EdgeDocument {

	return EdgeDocument{
		Value:         0,
		UpdateCounter: 0,
		Payload:       "",
		From:          colName + "/" + from,
		To:            colName + "/" + to,
	}
}

// createGraph creates a new graph.
// The operation is expected to succeed.
func (t *replication2Test) createGraph(graphName string,
	edgeCol string, fromCols []string, toCols []string, orphans []string,
	isSmart bool, isDisjoint bool, smartGraphAttribute string,
	satellites []string, numberOfShards int, replicatoinFactor int, writeConcern int) error {
	opts := struct {
		Name            string `json:"name"`
		EdgeDefinitions []struct {
			Collection string   `json:"collection"`
			From       []string `json:"from"`
			To         []string `json:"to"`
		} `json:"edgeDefinitions"`
		OrphanCollections []string `json:"orphanCollections,omitempty"`
		IsSmart           bool     `json:"isSmart"`
		IsDisjoint        bool     `json:"isDisjoint"`
		Options           struct {
			SmartGraphAttribute string   `json:"smartGraphAttribute,omitempty"`
			Satellites          []string `json:"satellites,omitempty"`
			NumberOfShards      int      `json:"numberOfShards,omitempty"`
			ReplicationFactor   int      `json:"replicationFactor,omitempty"`
			WriteConcern        int      `json:"writeConcern,omitempty"`
		}
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
		IsSmart:           isSmart,
		IsDisjoint:        isDisjoint,
		Options: struct {
			SmartGraphAttribute string   `json:"smartGraphAttribute,omitempty"`
			Satellites          []string `json:"satellites,omitempty"`
			NumberOfShards      int      `json:"numberOfShards,omitempty"`
			ReplicationFactor   int      `json:"replicationFactor,omitempty"`
			WriteConcern        int      `json:"writeConcern,omitempty"`
		}{
			SmartGraphAttribute: smartGraphAttribute,
			Satellites:          satellites,
			NumberOfShards:      numberOfShards,
			ReplicationFactor:   replicatoinFactor,
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
			[]int{400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		// 0, 503: recheck without erxpectations
		//     there: good
		//     not there: retry
		// 200   : good
		// 1, 500: collection couldn't be finished.
		//     there: failure
		//     not there: retry
		// 409   :
		//     first attempt: failure
		//     later attempts:
		//     recheck
		//         there: done
		//         else : failure

		if err[0] == nil {
			if resp[0].StatusCode == 200 {
				success = true
			} else {
				if resp[0].StatusCode == 1 || resp[0].StatusCode == 500 { // connection refused or not created
					checkRetry = true
					shouldNotExist = true
				} else if resp[0].StatusCode == 409 {
					if i == 1 {
						// This is a failure
						t.createGraphCounter.failed++
						t.reportFailure(test.NewFailure("Failed to create graph '%s': got 409 on first attempt", graphName))
						return maskAny(fmt.Errorf("Failed to create graph '%s': got 409 on first attempt", graphName))
					} else {
						shouldExist = true
					}
				}
				checkRetry = true
			}
		} else {
			// This is a failure
			t.createGraphCounter.failed++
			t.reportFailure(test.NewFailure("Failed to create graph '%s': %v", graphName, err[0]))
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
						t.reportFailure(test.NewFailure(
							"Failed to create graph '%s' rechecked and failed existence", graphName))
						return maskAny(fmt.Errorf("Failed to create  '%s' rechecked and failed existence", graphName))
					}
					success = true
				} else {
					if shouldExist {
						// This is a failure
						t.createGraphCounter.failed++
						t.reportFailure(test.NewFailure(
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
		test.NewFailure("Timed out while trying to create (%d) graph %s.", i, graphName))
	return maskAny(fmt.Errorf("Timed out while trying to create (%d) graph %s.", i, graphName))
}

func (t *replication2Test) graphExists(graphName string) (bool, error) {

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

func (t *replication2Test) dropGraph(graphName string, dropCollections bool) error {

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
			url, q, nil, []int{0, 1, 200, 404, 500, 503}, []int{400, 409, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d", resp[0].StatusCode, resp[0].Error_.ErrorNum)

		if err[0] != nil {
			// This is a failure
			t.dropGraphCounter.failed++
			t.reportFailure(test.NewFailure("Failed to drop graph '%s': %v", graphName, err[0]))
			return maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			// graph not found.
			// This can happen if the first attempt timed out, but did actually succeed.
			// So we accept this if there are multiple attempts.
			if i == 1 { // this is a failure in first run
				// Not enough attempts, this is a failure
				t.dropGraphCounter.failed++
				t.reportFailure(
					test.NewFailure("Failed to drop graph '%s': got 404 after only 1 attempt", graphName))
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
	t.reportFailure(test.NewFailure("Timed out (%d) while droping graph '%s'", i, graphName))
	return maskAny(fmt.Errorf("Timed out (%d) while droping graph '%s'", i, graphName))

}

func (t *replication2Test) createEdge() error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout)

	var seed = t.documentIdSeq
	t.documentIdSeq++
	from := strconv.FormatInt(t.existingDocSeeds[rand.Intn(len(t.existingDocSeeds))], 10)
	to := strconv.FormatInt(t.existingDocSeeds[rand.Intn(len(t.existingDocSeeds))], 10)
	document := NewEdgeDocument(from, to, t.docCollectionName)

	q := url.Values{}
	q.Set("waitForSync", "true")
	url := fmt.Sprintf("/_api/document/%s", t.edgeCollectionName)
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		checkRetry := false
		success := false

		t.log.Infof("Creating edge in collection '%s' with key %s...", t.edgeCollectionName)
		resp, err := t.client.Post(url, q, nil, document, "", nil,
			[]int{0, 1, 200, 201, 202, 409, 503}, []int{400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] == nil { // we have a response
			if resp[0].StatusCode == 503 || resp[0].StatusCode == 409 || resp[0].StatusCode == 0 {
				// 0, 503 and 409 -> check if accidentally successful
				checkRetry = true
			} else if resp[0].StatusCode != 1 {
				//FIXME: properly check for success
				success = true
			}
		} else { // failure
			t.singleDocCreateCounter.failed++
			t.reportFailure(
				test.NewFailure("Failed to create edge in collection '%s'", t.edgeCollectionName, err[0]))
			return maskAny(err[0])
		}

		//FIXME: implement checkretry - check if documents were still created even though we got a bad http response from coordinator
		if checkRetry {
			e := t.readExistingDocument(seed, false)
			success = e == nil
		}

		if success {
			t.existingDocSeeds = append(t.existingDocSeeds, seed)
			t.singleDocCreateCounter.succeeded++
			t.log.Infof("Creating document in '%s' succeeded", t.edgeCollectionName)
			return nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.singleDocCreateCounter.failed++
	t.reportFailure(
		test.NewFailure("Timed out while trying to create a document in '%s'.", t.edgeCollectionName))
	return maskAny(fmt.Errorf("Timed out while trying to create a document in '%s'.", t.edgeCollectionName))
}
