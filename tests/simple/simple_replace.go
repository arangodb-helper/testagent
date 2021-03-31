package simple

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

func (t *simpleTest) replaceExistingDocument(c *collection, key, rev string) (string, error) {
	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr, ifMatchStatus, explicitRev := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Replacing existing document '%s' (%s) in '%s' (name -> '%s')...", key, ifMatchStatus, c.name, newName)
	newDoc := UserDocument{
		Key:   key,
		Name:  fmt.Sprintf("Replaced named %s", key),
		Value: rand.Int(),
		Odd:   rand.Int()%2 == 0,
	}

	backoff := time.Millisecond * 250
	i := 0

	for true {

		i++
		if time.Now().After(testTimeout) {
			break;
		}

		checkRetry := false
		success := false
		update, err := t.client.Put(
			fmt.Sprintf("/_api/document/%s/%s", c.name, key), q, hdr, newDoc, "", nil,
			[]int{0, 200, 201, 202, 412, 503}, []int{400, 404}, operationTimeout, 1)

/*
 * 20x, if document was replaced
 * 404, if document did not exist
 * 412, if if-match was given and document already there
 * timeout, in which case the document might or might not exist
 *   connection refused with coordinator ==> simply try again with another
 *   connection error ("broken pipe") with coordinator, to be treated like
 *   a timeout
 * 503, cluster internal mishap, all bets off
 * Testagent:
 *   If first request gives correct result: OK
 *   if wrong result: ERROR  (include 503 in this case)
 *   if connection refused to coordinator: simply retry other
 *   if either timeout (or broken pipe with coordinator):
 * retry 5x and and ERROR then
*/

		if err == nil { // We have a response
			if update[0].StatusCode == 412 {
				if (!explicitRev && i > 1) || explicitRev {
					checkRetry = true
				} else {
					// We got a 412 without asking for an explicit revision on first attempt
					t.replaceExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(
							"Failed to replace existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							key, ifMatchStatus, c.name))
					return "", maskAny(
						fmt.Errorf(
							"Failed to replace existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match",
							key, ifMatchStatus, c.name))
				}
			} else if update[0].StatusCode == 503 || update[0].StatusCode == 0 {
				// 503 and 412 -> check if accidentally successful
				checkRetry = true
			} else {
				success = true
			}
		}

		if checkRetry {
			expected := c.existingDocs[key]
			expected.Name = newName
			d, e := readDocument(t, c.name, key, "", 240, true)

			if e == nil { // document does not exist
				if d.Equals(expected) {
					success = true
				} else {
					t.replaceExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(
							"Failed to update existing document '%s' (%s) in collection '%s': got 412 but has not been updated",
							key, ifMatchStatus, c.name))
					return "", maskAny(fmt.Errorf(
						"Failed to update existing document '%s' (%s) in collection '%s': got 412 but has not been updated",
						key, ifMatchStatus, c.name))
				}
			} else { // should never get here
				t.replaceExistingCounter.failed++
				t.reportFailure(
					test.NewFailure(
						"Failed to read existing document '%s' (%s) in collection '%s' that should have been updated: %v",
						key, ifMatchStatus, c.name, e))
				return "", maskAny(e)
			}
		}

		if success {
			// Update memory
			newDoc.rev = update[0].Rev
			c.existingDocs[key] = newDoc
			t.replaceExistingCounter.succeeded++
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

// replaceExistingDocumentWrongRevision replaces an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) replaceExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout := t.OperationTimeout
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr := ifMatchHeader(nil, rev)
	t.log.Infof("Replacing existing document '%s' wrong revision in '%s' (name -> '%s')...", key, collectionName, newName)
	newDoc := UserDocument{
		Key:   key,
		Name:  fmt.Sprintf("Replaced named %s", key),
		Value: rand.Int(),
		Odd:   rand.Int()%2 == 0,
	}
	_, err := t.client.Put(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, newDoc, "", nil, []int{412}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, 1)
	if err[0] != nil {
		// This is a failure
		t.replaceExistingWrongRevisionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to replace existing document '%s' wrong revision in collection '%s': %v", key, collectionName, err[0]))
		return maskAny(err[0])
	}
	t.replaceExistingWrongRevisionCounter.succeeded++
	t.log.Infof("Replacing existing document '%s' wrong revision in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}

// replaceNonExistingDocument replaces a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) replaceNonExistingDocument(collectionName string, key string) error {
	operationTimeout := t.OperationTimeout
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated non-existing name %s", time.Now())
	t.log.Infof("Replacing non-existing document '%s' in '%s' (name -> '%s')...", key, collectionName, newName)
	newDoc := UserDocument{
		Key:   key,
		Name:  fmt.Sprintf("Replaced named %s", key),
		Value: rand.Int(),
		Odd:   rand.Int()%2 == 0,
	}
	if _, err := t.client.Put(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, newDoc, "", nil, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, 1); err[0] != nil {
		// This is a failure
		t.replaceNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to replace non-existing document '%s' in collection '%s': %v", key, collectionName, err[0]))
		return maskAny(err[0])
	}
	t.replaceNonExistingCounter.succeeded++
	t.log.Infof("Replacing non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}
