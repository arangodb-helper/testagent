package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

func readDocument(t *simpleTest, col string, key string, rev string, seconds int, mustExist bool) (interface{}, error) {
	backoff := time.Millisecond * 100
	i := 0
	url := fmt.Sprintf("/_api/document/%s/%s", col, key)
	operationTimeout := time.Second * 30
	timeout := time.Now().Add(operationTimeout * 8)

	for  {
		i++
		if time.Now().After(timeout) {
			break;
		}
		hdr := ifMatchHeader(nil, rev)
		var result interface{}
		res, err := t.client.Get(url, nil, hdr, &result, []int{200, 201, 202, 404}, []int{400, 307}, operationTimeout, 1)

		if err[0] == nil {
			if res[0].StatusCode == 404 { // no such document
				if mustExist {
					t.reportFailure(
						test.NewFailure("Failed to read existing document '%s' (%s) in collection '%s': %v",
							key, col, err[0]))
					return nil, maskAny(err[0])
				} else {
					t.log.Errorf("Failed to read(%i) document %s in %s (&v).", i, key, col, err)
					return nil, nil
				}
			} else { // document found
				return result, nil
			}
		}

		time.Sleep(backoff)
		backoff += backoff
	}

	t.log.Errorf("Timed out while trying to read(%i) document %s in %s (&v).", i, key, col)
	return nil, maskAny(fmt.Errorf("Timed out"))

}

func (t *simpleTest) createDocument(c *collection, document interface{}, key string) (string, error) {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")

	t.log.Infof("Creating document '%s' in '%s'...", key, c.name)

	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break;
		}

		checkRetry := false
		success := false
		resp, err := t.client.Post(
			fmt.Sprintf("/_api/document/%s", c.name), q, nil, document, "", nil, []int{0, 200, 201, 202, 409, 503},
			[]int{400, 404, 307}, operationTimeout, 1)

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

	
		if err[0] == nil { // we have a response
			if resp[0].StatusCode == 503 || resp[0].StatusCode == 409 || resp[0].StatusCode == 0 {
				// 0, 503 and 409 -> check if accidentally successful
				checkRetry = true
			} else { // 20x
				success = true
			}
		} else { // failure
			t.reportFailure(
				test.NewFailure("Failed to create document '%s' (%s) in collection '%s': %v", key, c.name, err[0]))
			return "", maskAny(err[0])
		}

		if checkRetry {
			d, e := readDocument(t, c.name, key, "", 120, false)
			// replace == with Equals
			if e == nil && d == document {
				success = true
			}
		}

		if success {
			//c.existingDocs[key] = doc
			t.log.Infof("Creating document '%s' in '%s' succeeded", key, c.name)
			return resp[0].Rev, nil
		}

		t.log.Errorf("Failure %i to create existing document '%s' (%s) in collection '%s': got %i, retrying",
			i, key, c.name, resp[0].StatusCode)
		time.Sleep(backoff)
		backoff += backoff

	}

	// Overall timeout :(
	t.reportFailure(
		test.NewFailure("Timed out while trying to create(%i) document %s in %s.", i, key, c.name))
	return "", maskAny(fmt.Errorf("Timed out while trying to create(%i) document %s in %s.", i, key, c.name))

}
