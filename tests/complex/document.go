package complex

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
)

type TestDocument struct {
	Seed          int64  `json:"seed,omitempty"`
	Key           string `json:"_key,omitempty"`
	Rev           string `json:"_rev,omitempty"`
	UpdateCounter int    `json:"update_counter"`
}

func (t *TestDocument) equals(other *TestDocument) bool {
	return t.Key == other.Key && t.Seed == other.Seed && t.UpdateCounter == other.UpdateCounter
}

type BigDocument struct {
	TestDocument
	Value   int64  `json:"value"`
	Name    string `json:"name"`
	Odd     bool   `json:"odd"`
	Payload string `json:"payload"`
}

var (
	BackOffTime = time.Millisecond * 500 // to be overwritten in unittests only
)

func NewBigDocumentFromSeed(seed int64, payloadSize int) BigDocument {
	randGen := rand.New(rand.NewSource(seed))
	payloadBytes := make([]byte, payloadSize)
	lowerBound := 32
	upperBound := 126
	for i := 0; i < payloadSize; i++ {
		payloadBytes[i] = byte(randGen.Int31n(int32(upperBound-lowerBound)) + int32(lowerBound))
	}
	return BigDocument{
		TestDocument: TestDocument{Key: generateKeyFromSeed(seed),
			Seed:          seed,
			UpdateCounter: 0,
		},
		Value:   seed,
		Name:    strconv.FormatInt(seed, 10),
		Odd:     seed%2 == 1,
		Payload: string(payloadBytes),
	}
}

func NewBigDocumentWithName(seed int64, payloadSize int, name string) BigDocument {
	randGen := rand.New(rand.NewSource(seed))
	payloadBytes := make([]byte, payloadSize)
	lowerBound := 32
	upperBound := 126
	for i := 0; i < payloadSize; i++ {
		payloadBytes[i] = byte(randGen.Int31n(int32(upperBound-lowerBound)) + int32(lowerBound))
	}
	return BigDocument{
		TestDocument: TestDocument{Key: generateKeyFromSeed(seed),
			Seed:          seed,
			UpdateCounter: 0,
		},
		Value:   seed,
		Name:    name,
		Odd:     seed%2 == 1,
		Payload: string(payloadBytes),
	}
}

func NewBigDocumentFromTestDocument(testDocument TestDocument, payloadSize int) BigDocument {
	randGen := rand.New(rand.NewSource(testDocument.Seed))
	payloadBytes := make([]byte, payloadSize)
	lowerBound := 32
	upperBound := 126
	for i := 0; i < payloadSize; i++ {
		payloadBytes[i] = byte(randGen.Int31n(int32(upperBound-lowerBound)) + int32(lowerBound))
	}
	return BigDocument{
		TestDocument: testDocument,
		Value:        testDocument.Seed,
		Name:         testDocument.Key,
		Odd:          testDocument.Seed%2 == 1,
		Payload:      string(payloadBytes),
	}
}

func (t *ComplextTest) insertDocument(colName string, document any) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout)

	key := reflect.ValueOf(document).FieldByName("Key").String()

	q := url.Values{}
	q.Set("waitForSync", "true")
	url := fmt.Sprintf("/_api/document/%s", colName)
	backoff := BackOffTime
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
				success = true
			}
		} else { // failure
			t.singleDocCreateCounter.failed++
			t.reportFailure(
				test.NewFailure(t.Name(), "Failed to create document in collection '%s' with key %s: %v", colName, key, err[0]))
			return maskAny(err[0])
		}

		if checkRetry {
			exists, err := t.checkIfDocumentExists(colName, key)
			success = err == nil && exists
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
		test.NewFailure(t.Name(), "Timed out while trying to create a document in '%s with key %s'.", colName, key))
	return maskAny(fmt.Errorf("Timed out while trying to create a document in '%s with key %s'.", colName, key))
}

// checkIfDocumentExists checks if a document with given key exists in given collection
// The operation is expected to succeed.
func (t *ComplextTest) checkIfDocumentExists(colName string, key string) (bool, error) {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(t.OperationTimeout)
	i := 0
	url := fmt.Sprintf("/_api/document/%s/%s", colName, key)
	backoff := BackOffTime
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
		test.NewFailure(t.Name(),
			"Timed out reading document with key '%s' from collection '%s'", key, colName))
	return false, maskAny(fmt.Errorf("Timed out reading document with key '%s' from collection '%s'", key, colName))
}

// readExistingDocument reads an existing document.
// The operation is expected to succeed.
func (t *ComplextTest) readExistingDocument(colName string, expectedDocument BigDocument, skipExpectedValueCheck bool) error {

	operationTimeout := t.OperationTimeout / 5
	testTimeout := time.Now().Add(t.OperationTimeout)
	key := expectedDocument.Key
	resultPtr := &BigDocument{}
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
			url, nil, nil, resultPtr, []int{0, 1, 200, 503}, []int{400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] != nil {
			// This is a failure
			t.readExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(t.Name(),
					"Failed to read existing document with key '%s' from collection '%s': %v", key, colName, err[0]))
			return maskAny(err[0])
		} else {
			if resp[0].StatusCode == 200 {
				// Compare document against expected document
				if !skipExpectedValueCheck {
					if !resultPtr.Equals(expectedDocument) {
						// This is a failure
						t.readExistingCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(),
							"Read existing document with key '%s' from collection '%s' returned different values: got %v expected %v",
							key, colName, resultPtr, expectedDocument))
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
		test.NewFailure(t.Name(),
			"Timed out reading existing document with key '%s' from collection '%s'", key, colName))
	return maskAny(fmt.Errorf("Timed out reading existing document with key '%s' from collection '%s'", key, colName))
}

// readDocument tries to read a document. It retries up to `seconds` seconds,
// if timeout or connection refused or 503 happen, so these are never
// returned. If the document is not found (404), then this is considered
// to be an error, if `mustExist` is `true`, otherwise, the function
// simply returns `nil, nil`. In the good cases, `doc, nil` is returned.
// If the function times out, an error is returned. This function does
// not report failures.
func readDocument(t *ComplextTest, colName string, key string, rev string, seconds int, mustExist bool) (*BigDocument, error) {
	backoff := BackOffTime
	i := 0
	url := fmt.Sprintf("/_api/document/%s/%s", colName, key)
	operationTimeout := time.Duration(seconds/8) * time.Second
	timeout := time.Now().Add(time.Duration(seconds) * time.Second)

	for {
		i++
		if time.Now().After(timeout) {
			break
		}
		hdr := ifMatchHeader(nil, rev)
		var result BigDocument = BigDocument{}

		t.log.Infof(
			"Reading (%d) document '%s' (%s) in '%s' ...", i, key, rev, colName)
		res, err := t.client.Get(
			url, nil, hdr, &result, []int{0, 1, 200, 201, 202, 404, 503}, []int{400, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			res[0].StatusCode, res[0].Error_.ErrorNum, res[0].CoordinatorURL)

		if err[0] == nil {
			if res[0].StatusCode == 404 { // no such document
				if mustExist {
					t.readExistingCounter.failed++
					t.log.Errorf(
						"Failed to read(%d) existing document '%s' (%s) in collection '%s'",
						i, key, rev, colName)
					return nil, maskAny(fmt.Errorf("Failed to read(%d) existing document '%s' (%s) in collection '%s'",
						i, key, rev, colName))
				} else {
					t.log.Errorf("Failed to read(%d) document %s (%s) in %s got 404.", i, key, rev, colName)
					return nil, nil
				}
			} else if res[0].StatusCode >= 200 && res[0].StatusCode <= 202 { // document found
				t.readExistingCounter.succeeded++
				t.log.Infof(
					"Reading (%d) document '%s' (%s) in '%s' (name -> '%s') succeeded", i, key, rev, colName, result.Name)
				return &result, nil
			}
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}
	}

	t.readExistingCounter.failed++
	t.log.Errorf("Timed out while trying to read(%d) document %s in %s.", i, key, colName)
	return nil, maskAny(fmt.Errorf("Timed out while trying to read(%d) document %s in %s.", i, key, colName))

}

// readDocumentBySeed finds a single document by a custom unique field "seed"
func readDocumentBySeed(t *ComplextTest, colName string, seed int64) (*TestDocument, error) {
	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 5)
	backoff := BackOffTime
	i := 0

	var err []error
	var createResp []util.ArangoResponse
	var cursorResp CursorResponse

	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++

		t.log.Infof("Creating (%d) AQL query cursor for '%s'...", i, colName)
		queryReq := QueryRequest{
			Query:     fmt.Sprintf("FOR edge IN %s FILTER edge.seed==%d RETURN edge", colName, seed),
			BatchSize: 1,
			Count:     false,
		}

		createResp, err = t.client.Post(
			"/_api/cursor", nil, nil, queryReq, "", &cursorResp, []int{0, 1, 201, 410, 500, 503},
			[]int{200, 202, 307, 400, 404, 409}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			createResp[0].StatusCode, createResp[0].Error_.ErrorNum, createResp[0].CoordinatorURL)

		if err[0] != nil {
			// This is a failure
			t.queryCreateCursorCounter.failed++
			t.reportFailure(test.NewFailure(t.Name(), "Failed to create AQL cursor in collection '%s': %v", colName, err[0]))
			return nil, maskAny(err[0])
		} else if len(cursorResp.Result) == 0 {
			return nil, nil
		} else if len(cursorResp.Result) > 1 {
			return nil, errors.New("more than 1 document found")
		} else if createResp[0].StatusCode == 201 && len(cursorResp.Result) == 1 {
			t.queryCreateCursorCounter.succeeded++
			t.log.Infof("Creating AQL cursor for collection '%s' succeeded", colName)
			docMap := cursorResp.Result[0].(map[string]interface{})
			doc := TestDocument{
				Key:           docMap["_key"].(string),
				Rev:           docMap["_rev"].(string),
				Seed:          int64(docMap["seed"].(float64)),
				UpdateCounter: int(docMap["update_counter"].(float64)),
			}
			return &doc, nil
		}

		// Otherwise we fall through and simply try again. Note that if an
		// attempt times out it is OK to simply retry, even if the old one
		// eventually gets through, we can then simply work with the new
		// cursor. Furthermore note that we found that currently the undocumented
		// error code 500 can happen if a dbserver suffers from some chaos
		// during cursor creation. We can simply retry, too.
		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}
	}
	t.reportFailure(test.NewFailure(t.Name(),
		"Timed out while reading next AQL cursor batch in collection '%s' with same coordinator (%s)",
		colName, createResp[0].CoordinatorURL))
	return nil, maskAny(fmt.Errorf(
		"Timed out while reading next AQL cursor batch in collection '%s' with same coordinator (%s)",
		colName, createResp[0].CoordinatorURL))
}

// createRandomIfMatchHeader creates a request header with one of the following (randomly chosen):
// 1: with an `If-Match` entry for the given revision.
// 2: without an `If-Match` entry for the given revision.
// The bool response is true when an `If-Match` has been added, false otherwise.
func createRandomIfMatchHeader(hdr map[string]string, rev string) (map[string]string, string, bool) {
	if rev == "" {
		return hdr, "without If-Match", false
	}
	switch rand.Intn(2) {
	case 0:
		hdr = ifMatchHeader(hdr, rev)
		return hdr, "with If-Match", true
	default:
		return hdr, "without If-Match", false
	}
}

// ifMatchHeader creates a request header with an `If-Match` entry for the given revision.
func ifMatchHeader(hdr map[string]string, rev string) map[string]string {
	if hdr == nil {
		hdr = make(map[string]string)
	}
	if rev != "" {
		hdr["If-Match"] = rev
	}
	return hdr
}

// updateExistingDocument updates an existing document
func (t *ComplextTest) updateExistingDocument(colName string, oldDoc TestDocument) (*TestDocument, error) {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	url := fmt.Sprintf("/_api/document/%s/%s", colName, oldDoc.Key)

	hdr, ifMatchStatus, explicitRev := createRandomIfMatchHeader(nil, oldDoc.Rev)
	new_counter_value := oldDoc.UpdateCounter + 1
	delta := map[string]interface{}{
		"update_counter": new_counter_value,
	}
	backoff := BackOffTime
	i := 0

	for {
		i++
		if time.Now().After(testTimeout) {
			break
		}
		checkRetry := false
		success := false
		t.log.Infof(
			"Updating (%d) existing document '%s' (%s) in '%s' (update_counter -> '%d')...",
			i, oldDoc.Key, ifMatchStatus, colName, new_counter_value)
		update, err := t.client.Patch(url, q, hdr, delta, "", nil, []int{0, 1, 200, 201, 202, 409, 412, 503},
			[]int{400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			update[0].StatusCode, update[0].Error_.ErrorNum, update[0].CoordinatorURL)

		/**
		 *  20x, if document was replaced
		 *  400, if body bad
		 *  404, if document did not exist
		 *  409, if unique constraint would be violated (cannot happen here) or
		 *       if we had a timeout (or 503) and try again and we collide with
		 *       the previous request, in that case we need to checkRetry
		 *  412, if if-match was given and document already there
		 *  timeout, in which case the document might or might not have been written
		 *    after this, either one of the 5 things might have happened,
		 *    or nothing might have happened at all,
		 *    or any of this might still happen in the future
		 *  connection refused with coordinator ==> simply try again with another
		 *  connection error ("broken pipe") with coordinator, to be treated like
		 *    a timeout
		 *  503, cluster internal mishap, all bets off
		 *  If first request gives correct result: OK
		 *  if wrong result: ERROR
		 *  if connection refused to coordinator: simply retry other
		 *  if either timeout (or broken pipe with coordinator):
		 *    try to read the document repeatedly for up to 15s:
		 *      if new document there: treat as if op had worked
		 *    else (new version not appeared within 15s):
		 *      treat as if op has not worked and go to retry loop
		 */

		if err[0] == nil { // we have a response
			if update[0].StatusCode == 412 {
				if i == 1 || !explicitRev {
					// We got a 412 without asking for an explicit revision on first attempt
					t.updateExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(t.Name(),
							"Failed to update existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							oldDoc.Key, ifMatchStatus, colName))
					return nil, maskAny(
						fmt.Errorf(
							"Failed to update existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							oldDoc.Key, ifMatchStatus, colName))
				} else {
					checkRetry = true
				}
			} else if update[0].StatusCode == 0 || update[0].StatusCode == 409 || update[0].StatusCode == 503 {
				// 0, 409, 503 -> check if not accidentally successful
				t.log.Debugf(
					"Got status code %d. We need to re-check if the update operation was successfull.", update[0].StatusCode)
				checkRetry = true
			} else if update[0].StatusCode != 1 {
				oldDoc.Rev = update[0].Rev
				success = true
			}
		} else { // failure
			t.updateExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(t.Name(), "Failed to update existing document '%s' (%s) in collection '%s': %v",
					oldDoc.Key, ifMatchStatus, colName, err[0]))
			return nil, maskAny(
				fmt.Errorf(
					"Failed to update existing document '%s' (%s) in collection '%s': got unexpected code %d",
					oldDoc.Key, ifMatchStatus, colName, update[0].StatusCode))
		}

		if checkRetry {
			expected := oldDoc
			expected.UpdateCounter = new_counter_value
			t.log.Debugf(
				"Checking if the update operation was successfull. Expecting document: %v", expected)
			d, e := readDocument(t, colName, expected.Key, "", ReadTimeout, true)

			if e == nil { // document does not exist
				if d.TestDocument.equals(&expected) {
					oldDoc.Rev = d.Rev
					success = true
				} else if !d.TestDocument.equals(&oldDoc) {
					// If we see the existing one, we simply try again on the grounds
					// that the operation might not have happened. If it is still
					// happening, we might either collide or suddenly see the new
					// version.
					t.updateExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(t.Name(),
							"Failed to update existing document '%s' (%s) in collection '%s': found unexpected document: %v",
							oldDoc.Key, ifMatchStatus, colName, d))
					return nil, maskAny(fmt.Errorf(
						"Failed to update existing document '%s' (%s) in collection '%s': found unexpected document: %v",
						oldDoc.Key, ifMatchStatus, colName, d))
				} else if update[0].StatusCode == 412 {
					t.replaceExistingCounter.failed++
					t.reportFailure(test.NewFailure(t.Name(),
						"Failed to update existing document '%s' (%s) in collection '%s': found old document, and still got 412: %v",
						oldDoc.Key, ifMatchStatus, colName, d))
					return nil, maskAny(fmt.Errorf(
						"Failed to update existing document '%s' (%s) in collection '%s': found old document, and still got 412: %v",
						oldDoc.Key, ifMatchStatus, colName, d))

				}
			} else { // should never get here
				t.updateExistingCounter.failed++
				t.reportFailure(
					test.NewFailure(t.Name(),
						"Failed to read existing document '%s' (%s) in collection '%s' that should have been updated: %v",
						oldDoc.Key, ifMatchStatus, colName, e))
				return nil, maskAny(e)
			}
		}

		if success {
			// Update memory
			oldDoc.UpdateCounter = new_counter_value
			t.updateExistingCounter.succeeded++
			t.log.Infof(
				"Updating existing document '%s' (%s) in '%s' (update_counter -> '%d') succeeded",
				oldDoc.Key, ifMatchStatus, colName, new_counter_value)
			return &oldDoc, nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.updateExistingCounter.failed++
	t.reportFailure(
		test.NewFailure(t.Name(), "Timed out while trying to update(%d) document %s in %s.", i, oldDoc.Key, colName))
	return nil, maskAny(fmt.Errorf("Timed out while trying to update(%d) document %s in %s.", i, oldDoc.Key, colName))

}
