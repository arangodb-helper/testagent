package replication2

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

type TestDocument struct {
	Key string `json:"_key,omitempty"`
	Rev string `json:"_rev,omitempty"`
}

func (t *TestDocument) equals(other *TestDocument) bool {
	return t.Key == other.Key
}

func (t *Replication2Test) insertDocument(colName string, document any) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout)

	key := reflect.ValueOf(document).FieldByName("Key").String()

	q := url.Values{}
	q.Set("waitForSync", "true")
	url := fmt.Sprintf("/_api/document/%s", colName)
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		checkRetry := false
		success := false

		t.log.Infof("Creating document in collection '%s' with key %s...", colName, key)
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
				test.NewFailure("Failed to create document in collection '%s' with key %s", colName, key, err[0]))
			return maskAny(err[0])
		}

		if checkRetry {
			v, e := t.checkIfDocumentExists(colName, key)
			success = e == nil && v
		}

		if success {
			t.singleDocCreateCounter.succeeded++
			t.log.Infof("Creating document in '%s' with key %s succeeded", colName, key)
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
		test.NewFailure("Timed out while trying to create a document in '%s with key %s'.", colName, key))
	return maskAny(fmt.Errorf("Timed out while trying to create a document in '%s with key'.", colName, key))
}

//FIXME: this method currently can't work properly. we need to extend the client to support bulk requests,
// which may return an array of errors instead of just one ArangoError object

// createDocuments creates a new documents in bulk
// func (t *Replication2Test) createDocuments(numberOfDocuments int, startValue int64) error {

// 	operationTimeout := t.OperationTimeout
// 	testTimeout := time.Now().Add(operationTimeout * 4)

// 	q := url.Values{}
// 	q.Set("waitForSync", "true")
// 	url := fmt.Sprintf("/_api/document/%s", t.docCollectionName)
// 	backoff := time.Millisecond * 250
// 	i := 0

// 	for {

// 		i++
// 		if time.Now().After(testTimeout) {
// 			break
// 		}

// 		checkRetry := false
// 		success := false

// 		var documents []TestDocument
// 		documents = make([]TestDocument, numberOfDocuments)
// 		for i := startValue; i < startValue+int64(numberOfDocuments); i++ {
// 			documents = append(documents, NewTestDocument(i, t.DocumentSize))
// 		}
// 		t.log.Infof("Creating (%d) documents in collection '%s'...", numberOfDocuments, t.docCollectionName)
// 		resp, err := t.client.Post(url, q, nil, documents, "", nil,
// 			[]int{0, 1, 200, 201, 202, 409, 503}, []int{400, 404, 307}, operationTimeout, 1)
// 		t.log.Infof("... got http %d - arangodb %d via %s",
// 			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

// 		if err[0] == nil { // we have a response
// 			if resp[0].StatusCode == 503 || resp[0].StatusCode == 409 || resp[0].StatusCode == 0 {
// 				// 0, 503 and 409 -> check if accidentally successful
// 				checkRetry = true
// 			} else if resp[0].StatusCode != 1 {
// 				//FIXME: properly check for success
// 				success = true
// 			}
// 		} else { // failure
// 			t.bulkCreateCounter.failed++
// 			t.reportFailure(
// 				test.NewFailure("Failed to create %d documents in collection '%s'", numberOfDocuments, t.docCollectionName, err[0]))
// 			return maskAny(err[0])
// 		}

// 		//FIXME: implement checkretry - check if documents were still created even though we got a bad http response from coordinator
// 		if checkRetry {
// 			// 	d, e := readDocument(t, c.name, key, "", ReadTimeout, false)
// 			// 	// replace == with Equals
// 			// 	if e == nil && d != nil && d.Equals(document) {
// 			// 		document.Rev = d.Rev
// 			// 		success = true
// 			// 	}
// 		}

// 		if success {
// 			for _, v := range documents {
// 				t.existingDocSeeds = append(t.existingDocSeeds, v.Value)
// 			}
// 			t.bulkCreateCounter.succeeded++
// 			t.log.Infof("Creating %d documents in '%s' succeeded", numberOfDocuments, t.docCollectionName)
// 			return nil
// 		}

// 		time.Sleep(backoff)
// 		if backoff < time.Second*5 {
// 			backoff += backoff
// 		}

// 	}

// 	// Overall timeout :(
// 	t.bulkCreateCounter.failed++
// 	t.reportFailure(
// 		test.NewFailure("Timed out while trying to create %d documents in '%s'.", numberOfDocuments, t.docCollectionName))
// 	return maskAny(fmt.Errorf("Timed out while trying to create %d documents in '%s'.", numberOfDocuments, t.docCollectionName))
// }

// checkIfDocumentExists checks if a document with given key exists in given collection
// The operation is expected to succeed.
func (t *Replication2Test) checkIfDocumentExists(colName string, key string) (bool, error) {

	operationTimeout := t.OperationTimeout / 5
	testTimeout := time.Now().Add(t.OperationTimeout)
	i := 0
	url := fmt.Sprintf("/_api/document/%s/%s", colName, key)
	backoff := time.Millisecond * 100
	var result TestDocument

	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++

		t.log.Infof("Reading existing document with key '%s' from collection '%s'...", key, colName)
		resp, err := t.client.Get(
			url, nil, nil, &result, []int{0, 1, 200, 404, 503}, []int{400, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] != nil {
			return false, maskAny(err[0])
		} else {
			if resp[0].StatusCode == 200 {
				return true, nil
			} else if resp[0].StatusCode == 404 {
				return false, nil
			}
		}
		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	t.reportFailure(
		test.NewFailure(
			"Timed out reading document with key '%s' from collection '%s'", key, colName))
	return false, maskAny(fmt.Errorf("Timed out reading document with key '%s' from collection '%s'", key, colName))
}

// readExistingDocument reads an existing document.
// The operation is expected to succeed.
func (t *Replication2Test) readExistingDocument(colName string, expectedDocument any, skipExpectedValueCheck bool) error {

	operationTimeout := t.OperationTimeout / 5
	testTimeout := time.Now().Add(t.OperationTimeout)
	key := reflect.ValueOf(expectedDocument).FieldByName("Key").String()
	result := reflect.New(reflect.TypeOf(expectedDocument)).Interface()
	i := 0
	url := fmt.Sprintf("/_api/document/%s/%s", colName, key)
	backoff := time.Millisecond * 100

	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++

		t.log.Infof("Reading existing document with key '%s' from collection '%s'...", key, colName)
		resp, err := t.client.Get(
			url, nil, nil, &result, []int{0, 1, 200, 503}, []int{400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] != nil {
			// This is a failure
			t.readExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(
					"Failed to read existing document with key '%s' from collection '%s': %v", key, colName, err[0]))
			return maskAny(err[0])
		} else {
			if resp[0].StatusCode == 200 {
				// Compare document against expected document
				if !skipExpectedValueCheck {
					if !reflect.ValueOf(result).MethodByName("Equals").Call([]reflect.Value{reflect.ValueOf(expectedDocument)})[0].Interface().(bool) {
						// This is a failure
						t.readExistingCounter.failed++
						t.reportFailure(test.NewFailure(
							"Read existing document with key '%s' from collection '%s' returned different values: got %v expected %v",
							key, colName, result, expectedDocument))
						return maskAny(fmt.Errorf("Read returned invalid values"))
					}
				}
				t.readExistingCounter.succeeded++
				t.log.Infof("Reading existing document with key '%s' from collection '%s' succeeded", key, colName)
				return nil
			}
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	t.readExistingCounter.failed++
	t.reportFailure(
		test.NewFailure(
			"Timed out reading existing document with key '%s' from collection '%s'", key, colName))
	return maskAny(fmt.Errorf("Timed out reading existing document with key '%s' from collection '%s'", key, colName))
}
