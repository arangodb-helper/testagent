package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

// updateExistingDocument updates an existing document
func (t *simpleTest) updateExistingDocument(c *collection, key, rev string) (string, error) {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	url := fmt.Sprintf("/_api/document/%s/%s", c.name, key)

	hdr, ifMatchStatus, explicitRev := createRandomIfMatchHeader(nil, rev)
	delta := map[string]interface{}{
		"name": newName,
	}
	doc := c.existingDocs[key]
	backoff := time.Millisecond * 250
	i := 0

	for {
		i++
		if time.Now().After(testTimeout) {
			break
		}

		checkRetry := false
		success := false
		t.log.Infof(
			"Updating (%d) existing document '%s' (%s) in '%s' (name -> '%s')...",
			i, key, ifMatchStatus, c.name, newName)
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
			} else if update[0].StatusCode == 0 || update[0].StatusCode == 409 || update[0].StatusCode == 503 {
				// 0, 409, 503 -> check if not accidentally successful
				checkRetry = true
			} else if update[0].StatusCode != 1 {
				success = true
			}
		} else { // failure
			t.updateExistingCounter.failed++
			t.reportFailure(
				test.NewFailure("Failed to update existing document '%s' (%s) in collection '%s': %v",
					key, ifMatchStatus, c.name, err[0]))
			return "", maskAny(err[0])
		}

		if checkRetry {
			expected := c.existingDocs[key]
			expected.Name = newName
			d, e := readDocument(t, c.name, key, "", 128, true)

			if e == nil { // document does not exist
				if d.Equals(expected) {
					success = true
				} else if !d.Equals(doc) {
					// If we see the existing one, we simply try again on the grounds
					// that the operation might not have happened. If it is still
					// happening, we might either collide or suddenly see the new
					// version.
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
				"Updating existing document '%s' (%s) in '%s' (name -> '%s') succeeded",
				key, ifMatchStatus, c.name, newName)
			return update[0].Rev, nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.updateExistingCounter.failed++
	t.planCollectionDrop(c.name)
	t.reportFailure(
		test.NewFailure("Timed out while trying to update(%d) document %s in %s.", i, key, c.name))
	return "", maskAny(fmt.Errorf("Timed out while trying to update(%d) document %s in %s.", i, key, c.name))

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
	delta := map[string]interface{}{
		"name": newName,
	}
	url := fmt.Sprintf("/_api/document/%s/%s", collectionName, key)
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		t.log.Infof(
			"Updating (%d) existing document '%s' wrong revision in '%s' (name -> '%s')...",
			i, key, collectionName, newName)
		resp, err := t.client.Patch(url, q, hdr, delta, "", nil, []int{0, 1, 412, 503},
			[]int{200, 201, 202, 400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

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
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	t.updateExistingWrongRevisionCounter.failed++
	t.reportFailure(
		test.NewFailure(
			"Timed out while updating (%d) existing document '%s' wrong revision in collection '%s'",
			i, key, collectionName))
	return maskAny(
		fmt.Errorf("Timed out while updating (%d) existing document '%s' wrong revision in collection '%s'",
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
	delta := map[string]interface{}{
		"name": newName,
	}
	url := fmt.Sprintf("/_api/document/%s/%s", collectionName, key)
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		t.log.Infof(
			"Updating (%d) non-existing document '%s' in '%s' (name -> '%s')...", i, key, collectionName, newName)
		resp, err := t.client.Patch(url, q, nil, delta, "", nil, []int{0, 1, 404, 503},
			[]int{200, 201, 202, 400, 412, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

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
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	t.updateNonExistingCounter.failed++
	t.reportFailure(
		test.NewFailure(
			"Timeout while updating (%d) non-existing document '%s' in collection '%s'", i, key, collectionName))
	return maskAny(
		fmt.Errorf(
			"Timeout while updating (%d) non-existing document '%s' in collection '%s'", i, key, collectionName))

}
