package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

func readDocument(t *simpleTest, col string, key string, rev string, seconds int, mustExist bool) (*UserDocument, error) {
	backoff := time.Millisecond * 100
	i := 0
	url := fmt.Sprintf("/_api/document/%s/%s", col, key)
	operationTimeout := time.Duration(seconds / 8) * time.Second
	timeout := time.Now().Add(time.Duration(seconds) * time.Second)

	for  {
		i++
		if time.Now().After(timeout) {
			break;
		}
		hdr := ifMatchHeader(nil, rev)
		var result *UserDocument

		t.log.Infof(
			"Reading (%d) document '%s' (%s) in '%s' ...", i, key, rev, col)
		res, err := t.client.Get(
			url, nil, hdr, &result, []int{0, 1, 200, 201, 202, 404, 503}, []int{400, 307}, operationTimeout, 1)

		if err[0] == nil {
			if res[0].StatusCode == 404 { // no such document
				if mustExist {
					t.readExistingCounter.failed++
					t.reportFailure(
						test.NewFailure("Failed to read(%d) existing document '%s' (%s) in collection '%s': %v",
							i, key, rev, col, err[0]))
					return nil, maskAny(err[0])
				} else {
					t.log.Errorf("Failed to read(%d) document %s (%s) in %s (&v).", i, key, rev, col, err)
					return nil, nil
				}
			} else if res[0].StatusCode >= 200 && res[0].StatusCode <= 202 { // document found
				t.readExistingCounter.succeeded++
				t.log.Infof(
					"Reading (%d) document '%s' (%s) in '%s' (name -> '%s') succeeded", i, key, rev, col)
				return result, nil
			} 
		}

		t.log.Infof(
			"Reading (%d) document '%s' (%s) in '%s' got %d", i, key, rev, col, res[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}
	}

	t.readExistingCounter.failed++
	t.log.Errorf("Timed out while trying to read(%d) document %s in %s (&v).", i, key, col)
	return nil, maskAny(fmt.Errorf("Timed out while trying to read(%d) document %s in %s (&v).", i, key, col))

}

/*
	POST /_api/document
	{"body":1}
	-->
	20x, if document was inserted
	400, if body bad
	404, if collection does not exist
	409, if document already existed
	timeout, in which case the document might or might not have been written
		after this, either one of the 4 things might have happened,
		or nothing might have happened at all,
		or any of this might still happen in the future
	connection refused with coordinator ==> simply try again with another
	connection error ("broken pipe") with coordinator, to be treated like
		a timeout
	503, cluster internal mishap, all bets off
	If first request gives correct result: OK
	if wrong result: ERROR  (include 503 in this case)
	if connection refused to coordinator: simply retry other
	if either timeout (or broken pipe with coordinator):
		try to read the document repeatedly for up to 15s:
			if document there: treat as if insert had worked
		else (not appeared within 15s):
			treat as if insert has not worked and go to retry loop
*/

// createDocument creates a new document.
// The operation is expected to succeed.
func (t *simpleTest) createDocument(c *collection, document UserDocument, key string) (string, error) {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	url := fmt.Sprintf("/_api/document/%s", c.name)
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break;
		}

		checkRetry := false
		success := false

		t.log.Infof("Creating (%d) document '%s' in '%s'...", i, key, c.name)
		resp, err := t.client.Post(url, q, nil, document, "", nil,
			[]int{0, 1, 200, 201, 202, 409, 503},	[]int{400, 404, 307}, operationTimeout, 1)

		if err[0] == nil { // we have a response
			if resp[0].StatusCode == 503 || resp[0].StatusCode == 409 || resp[0].StatusCode == 0 {
				// 0, 503 and 409 -> check if accidentally successful
				checkRetry = true
			} else if resp[0].StatusCode != 1 {
				success = true
			}
		} else { // failure
			t.createCounter.failed++
			t.reportFailure(
				test.NewFailure("Failed to create document '%s' in collection '%s': %v", key, c.name, err[0]))
			return "", maskAny(err[0])
		}

		if checkRetry {
			d, e := readDocument(t, c.name, key, "", 128, false)
			// replace == with Equals
			if e == nil && d != nil && d.Equals(document) {
				success = true
			}
		}

		if success {
			//c.existingDocs[key] = doc
			t.createCounter.succeeded++
			t.log.Infof("Creating document '%s' in '%s' succeeded", key, c.name)
			return resp[0].Rev, nil
		}

		t.log.Infof("Creating (%d) document '%s' in '%s' got %d", i, key, c.name, resp[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.createCounter.failed++
	t.reportFailure(
		test.NewFailure("Timed out while trying to create(%d) document %s in %s.", i, key, c.name))
	return "", maskAny(fmt.Errorf("Timed out while trying to create(%d) document %s in %s.", i, key, c.name))

}
