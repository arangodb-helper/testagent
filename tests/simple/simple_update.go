package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

func (t *simpleTest) updateExistingDocument(c *collection, key, rev string) (string, error) {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())

	hdr, ifMatchStatus, explicitRev := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Updating existing document '%s' (%s) in '%s' (name -> '%s')...", key, ifMatchStatus, c.name, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	doc := c.existingDocs[key]
	backoff := time.Millisecond * 250
	i := 0

	for  {
		i++
		if time.Now().After(testTimeout) {
			break;
		}

		checkRetry := false
		success := false
		update, err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", c.name, key), q,
			hdr, delta, "", nil, []int{0, 200, 201, 202, 412, 503}, []int{400, 404, 409, 307}, operationTimeout, 1)

/**
 *  20x, if document was replaced
 *  400, if body bad
 *  404, if document did not exist
 *  409, if unique constraint would be violated
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
 *  if wrong result: ERROR  (include 503 in this case)
 *  if connection refused to coordinator: simply retry other
 *  if either timeout (or broken pipe with coordinator):
 *    try to read the document repeatedly for up to 15s:
 *      if new document there: treat as if op had worked
 *    else (new version not appeared within 15s):
 *      treat as if op has not worked and go to retry loop
 */

		if err[0] == nil { // we have a response
			if update[0].StatusCode == 412 {
				if (!explicitRev && i > 1) || explicitRev {
					checkRetry = true
				} else {
					// We got a 412 without asking for an explicit revision on first attempt
					t.updateExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(
							"Failed to update existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							key, ifMatchStatus, c.name))
					return "", maskAny(
						fmt.Errorf(
							"Failed to update existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							key, ifMatchStatus, c.name))
				}
			} else if update[0].StatusCode == 503 || update[0].StatusCode == 0 {
				// 503 and 412 -> check if accidentally successful
				checkRetry = true
			} else {
				success = true
			}
		}	else { // failure
			t.updateExistingCounter.failed++
			t.reportFailure(
				test.NewFailure("Failed to update existing document '%s' (%s) in collection '%s': %v",
					key, ifMatchStatus, c.name, err[0]))
			return "", maskAny(err[0])
		}

		if checkRetry {
			expected := c.existingDocs[key]
			expected.Name = newName
			d, e := readDocument(t, c.name, key, "", 240, true)

			if e == nil { // document does not exist
				if d.Equals(expected) {
					success = true
				} else {
					t.updateExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(
							"Failed to update existing document '%s' (%s) in collection '%s': got 412 but has not been updated",
							key, ifMatchStatus, c.name))
					return "", maskAny(fmt.Errorf(
						"Failed to update existing document '%s' (%s) in collection '%s': got 412 but has not been updated",
						key, ifMatchStatus, c.name))
				}
			} else { // should never get here
				t.updateExistingCounter.failed++
				t.reportFailure(
					test.NewFailure(
						"Failed to read existing document '%s' (%s) in collection '%s' that should have been updated: %v",
						key, ifMatchStatus, c.name, e))
				return "", maskAny(e)
			}
		}

		if success {
			// Update memory
			doc.Name = newName
			doc.rev = update[0].Rev
			c.existingDocs[key] = doc
			t.updateExistingCounter.succeeded++
			t.log.Infof(
				"Updating existing document '%s' (%s) in '%s' (name -> '%s') succeeded", key, ifMatchStatus, c.name, newName)
			return update[0].Rev, nil
		}

		t.log.Errorf("Failure %i to update existing document '%s' (%s) in collection '%s': got %i, retrying",
			i, key, c.name, update[0].StatusCode)
		time.Sleep(backoff)
		backoff += backoff

	}

	// Overall timeout :(
	t.reportFailure(
		test.NewFailure("Timed out while trying to update(%i) document %s in %s.", i, key, c.name))
	return "", maskAny(fmt.Errorf("Timed out while trying to update(%i) document %s in %s.", i, key, c.name))

}

// updateExistingDocumentWrongRevision updates an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) updateExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr := ifMatchHeader(nil, rev)
	t.log.Infof(
		"Updating existing document '%s' wrong revision in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break;
		}

		resp, err := t.client.Patch(
			fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, delta, "", nil,
			[]int{0, 1, 412, 503}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, 1)

		if err[0] == nil {
			if resp[0].StatusCode == 412 {
				t.updateExistingWrongRevisionCounter.succeeded++
				t.log.Infof(
					"Updating existing document '%s' wrong revision in '%s' (name -> '%s') succeeded",
					key, collectionName, newName)
				return nil
			}
		} else {
			// This is a failure
			t.updateExistingWrongRevisionCounter.failed++
			t.reportFailure(
				test.NewFailure(
					"Failed to update existing document '%s' wrong revision in collection '%s': %v",
					key, collectionName, err[0]))
			return maskAny(err[0])
		}

		time.Sleep(backoff)
		backoff += backoff
	}

	t.reportFailure(
		test.NewFailure(
			"Timed out while updating (%i) existing document '%s' wrong revision in collection '%s'",
			i, key, collectionName))
	return maskAny(
		fmt.Errorf("Timed out while updating (%i) existing document '%s' wrong revision in collection '%s'",
			i, key, collectionName))
		
}

// updateNonExistingDocument updates a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) updateNonExistingDocument(collectionName string, key string) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated non-existing name %s", time.Now())
	t.log.Infof("Updating non-existing document '%s' in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}

	backoff := time.Millisecond * 250
	i := 0
	
	for {

		i++
		if time.Now().After(testTimeout) {
			break;
		}
		
		resp, err := t.client.Patch(
			fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, delta, "", nil,
			[]int{0, 1, 404, 503}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, 1)

		if err[0] == nil {
			if resp[0].StatusCode == 404 {
				t.updateNonExistingCounter.succeeded++
				t.log.Infof(
					"Updating non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
				return nil
			}
		} else {
			// This is a failure
			t.updateNonExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(
					"Failed to update non-existing document '%s' in collection '%s': %v", key, collectionName, err[0]))
			return maskAny(err[0])
		}

		time.Sleep(backoff)
		backoff += backoff
	}

	t.reportFailure(
		test.NewFailure(
			"Timeout while updating (%i) non-existing document '%s' in collection '%s'", i, key, collectionName))
	return maskAny(
		fmt.Errorf(
			"Timeout while updating (%i) non-existing document '%s' in collection '%s'", i, key, collectionName))
	
}
