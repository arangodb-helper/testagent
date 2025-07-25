package simple

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

// replaceExistingDocument tries to replace an existing document.
func (t *simpleTest) replaceExistingDocument(c *collection, key, rev string) (string, error) {
	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr, ifMatchStatus, explicitRev := createRandomIfMatchHeader(nil, rev)
	newDoc := UserDocument{
		Key:   key,
		Name:  fmt.Sprintf("Replaced named %s", key),
		Value: rand.Int(),
		Odd:   rand.Int()%2 == 0,
	}
	url := fmt.Sprintf("/_api/document/%s/%s", c.name, key)

	backoff := time.Millisecond * 250
	i := 0

	mustNotBeReplacedYet := true
	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		checkRetry := false
		success := false
		t.log.Infof("Replacing (%d) existing document '%s' (%s) in '%s' (name -> '%s')...",
			i, key, ifMatchStatus, c.name, newName)
		update, err := t.client.Put(
			url, q, hdr, newDoc, "", nil, []int{0, 1, 200, 201, 202, 409, 410, 412, 503},
			[]int{400, 404}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			update[0].StatusCode, update[0].Error_.ErrorNum, update[0].CoordinatorURL)

		/*
		 * 20x, if document was replaced
		 * 404, if document did not exist
		 * 410, if operation could not been completed, nothing was written
		 * 412, if if-match was given and document already there
		 * timeout, in which case the document might or might not exist
		 *   connection refused with coordinator ==> simply try again with another
		 *   connection error ("broken pipe") with coordinator, to be treated like
		 *   a timeout
		 * 503, cluster internal mishap, all bets off
		 * Testagent:
		 *   If first request gives correct result: OK
		 *   if wrong result: ERROR
		 *   if connection refused to coordinator: simply retry other
		 *   if either timeout (or broken pipe with coordinator):
		 *     try to read the document repeatedly
		 *     if new document there:
		 *       success!
		 *     if old document there:
		 *       treat it as if the replace has not worked
		 */

		if err[0] == nil { // We have a response
			if update[0].StatusCode == 412 {
				if i == 1 || !explicitRev {
					// We got a 412 without asking for an explicit revision on first attempt
					t.replaceExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(t.Name(),
							"Failed to replace existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							key, ifMatchStatus, c.name))
					return "", maskAny(
						fmt.Errorf(
							"Failed to replace existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							key, ifMatchStatus, c.name))
				} else {
					checkRetry = true
					mustNotBeReplacedYet = false
				}
			} else if update[0].StatusCode == 503 || update[0].StatusCode == 409 ||
				update[0].StatusCode == 0 {
				// 503 and 0 (timeout) -> check if accidentally successful
				// 409 can happen if we had a previous timeout and the previous
				// request and the current one collide with their write transactions.
				// In this case we want to check if any of the operations was
				// successful in changing the document.
				checkRetry = true
				mustNotBeReplacedYet = false
			} else if update[0].StatusCode == 410 {
				checkRetry = true
			} else if update[0].StatusCode != 1 {
				newDoc.Rev = update[0].Rev
				success = true
				mustNotBeReplacedYet = false
			}
		} else {
			t.replaceExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(t.Name(),
					"Failed to replace existing document '%s' (%s) in collection '%s': got unexpected code %d",
					key, ifMatchStatus, c.name, update[0].StatusCode))
			return "", maskAny(
				fmt.Errorf(
					"Failed to replace existing document '%s' (%s) in collection '%s': got unexpected code %d",
					key, ifMatchStatus, c.name, update[0].StatusCode))
		}

		if checkRetry {
			expected := c.existingDocs[key]
			d, e := readDocument(t, c.name, key, "", ReadTimeout, true)

			if e == nil {

				if d.Equals(newDoc) {
					if !mustNotBeReplacedYet {
						newDoc.Rev = d.Rev
						success = true
					} else {
						t.replaceExistingCounter.failed++
						t.reportFailure(test.NewFailure(t.Name(),
							"Failed to replace existing document '%s' (%s) in collection '%s': document was replaced unexpectedly. Last returned HTTP status code: %d",
							key, ifMatchStatus, c.name, update[0].StatusCode))
						return "", maskAny(fmt.Errorf(
							"Failed to replace existing document '%s' (%s) in collection '%s': document was replaced unexpectedly. Last returned HTTP status code: %d",
							key, ifMatchStatus, c.name, update[0].StatusCode))
					}
				} else if !d.Equals(expected) {
					t.replaceExistingCounter.failed++
					t.reportFailure(test.NewFailure(t.Name(),
						"Failed to replace existing document '%s' (%s) in collection '%s': found unexpected document: %v",
						key, ifMatchStatus, c.name, d))
					return "", maskAny(fmt.Errorf(
						"Failed to replace existing document '%s' (%s) in collection '%s': found unexpected document: %v",
						key, ifMatchStatus, c.name, d))
				} else if update[0].StatusCode == 412 {
					t.replaceExistingCounter.failed++
					t.reportFailure(test.NewFailure(t.Name(),
						"Failed to replace existing document '%s' (%s) in collection '%s': found old document, and still got 412: %v",
						key, ifMatchStatus, c.name, d))
					return "", maskAny(fmt.Errorf(
						"Failed to replace existing document '%s' (%s) in collection '%s': found old document, and still got 412: %v",
						key, ifMatchStatus, c.name, d))

				}
			} else { // should never get here
				t.replaceExistingCounter.failed++
				t.reportFailure(test.NewFailure(t.Name(),
					"Failed to read existing document '%s' (%s) in collection '%s' that should have been updated: %v",
					key, ifMatchStatus, c.name, e))
				return "", maskAny(e)
			}
		}

		if success {
			// Update memory
			c.existingDocs[key] = newDoc
			t.replaceExistingCounter.succeeded++
			t.log.Infof(
				"Replacing existing document '%s' (%s) in '%s' (name -> '%s') succeeded",
				key, ifMatchStatus, c.name, newName)
			return newDoc.Rev, nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}
	}

	// Overall timeout :(
	t.replaceExistingCounter.failed++
	t.planCollectionDrop(c.name)
	t.reportFailure(test.NewFailure(t.Name(), "Timed out (%d) while trying to replace document %s in %s.", i, key, c.name))
	return "", maskAny(fmt.Errorf("Timed out (%d) while trying to replace document %s in %s.", i, key, c.name))

}

// replaceExistingDocumentWrongRevision replaces an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) replaceExistingDocumentWrongRevision(collectionName string, key, rev string) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr := ifMatchHeader(nil, rev)
	newDoc := UserDocument{
		Key:   key,
		Name:  fmt.Sprintf("Replaced named %s", key),
		Value: rand.Int(),
		Odd:   rand.Int()%2 == 0,
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
			"Replacing (%d) existing document '%s' wrong revision in '%s' (name -> '%s')...",
			i, key, collectionName, newName)
		resp, err := t.client.Put(
			url, q, hdr, newDoc, "", nil, []int{0, 1, 410, 412, 503},
			[]int{200, 201, 202, 400, 404, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] == nil {
			if resp[0].StatusCode == 412 {
				t.replaceExistingWrongRevisionCounter.succeeded++
				t.log.Infof("Replacing existing document '%s' wrong revision in '%s' (name -> '%s') succeeded", key, collectionName, newName)
				return nil
			}
			// In cases 0, 1, 410 and 503, we fall through here and try again
		} else {
			// This is a failure
			t.replaceExistingWrongRevisionCounter.failed++
			t.reportFailure(
				test.NewFailure(t.Name(),
					"Failed to replace existing document '%s' wrong revision in collection '%s': %v",
					key, collectionName, err[0]))
			return maskAny(err[0])
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}
	}

	t.replaceExistingWrongRevisionCounter.failed++
	t.reportFailure(
		test.NewFailure(t.Name(),
			t.Name(), "Timed out while replacing (%d) existing document '%s' wrong revision in collection '%s'",
			i, key, collectionName))
	return maskAny(
		fmt.Errorf(
			"Timed out while replacing (%d) existing document '%s' wrong revision in collection '%s'",
			i, key, collectionName))

}

// replaceNonExistingDocument replaces a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) replaceNonExistingDocument(collectionName string, key string) error {
	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated non-existing name %s", time.Now())
	newDoc := UserDocument{
		Key:   key,
		Name:  fmt.Sprintf("Replaced named %s", key),
		Value: rand.Int(),
		Odd:   rand.Int()%2 == 0,
	}

	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		t.log.Infof("Replacing (%d) non-existing document '%s' in '%s' (name -> '%s')...",
			i, key, collectionName, newName)
		resp, err := t.client.Put(
			fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, newDoc, "", nil,
			[]int{0, 1, 404, 410, 503}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] == nil {
			if resp[0].StatusCode == 404 {
				t.replaceNonExistingCounter.succeeded++
				t.log.Infof(
					"Replacing non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
				return nil
			}
			// In cases 0, 1, 410 and 503 we fall through here and try again.
		} else {
			// This is a failure
			t.replaceNonExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(t.Name(),
					"Failed to replace non-existing document '%s' in collection '%s': %v", key, collectionName, err[0]))
			return maskAny(err[0])
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}
	}

	t.replaceNonExistingCounter.failed++
	t.reportFailure(
		test.NewFailure(t.Name(),
			t.Name(), "Timeout while replacing (%d) non-existing document '%s' in collection '%s'", i, key, collectionName))
	return maskAny(
		fmt.Errorf(
			"Timeout while replacing (%d) non-existing document '%s' in collection '%s'", i, key, collectionName))

}
