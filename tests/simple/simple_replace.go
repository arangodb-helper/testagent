package simple

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/arangodb/testAgent/service/test"
)

// replaceExistingDocument replaces an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) replaceExistingDocument(c *collection, key, rev string) (string, error) {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
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
	update, err := t.client.Put(fmt.Sprintf("/_api/document/%s/%s", c.name, key), q, hdr, newDoc, "", nil, []int{200, 201, 202, 412}, []int{400, 404, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.replaceExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to replace existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, c.name, err))
		return "", maskAny(err)
	} else if update.StatusCode == 412 {
		if explicitRev {
			// Expected revision did NOT match.
			// This may happen when a coordinator succeeds in the first attempt but we've already timed out.
			// Check document against what we expect after the replace.
			expected := newDoc
			if match, rev, err := t.isDocumentEqualTo(c, key, expected); err != nil {
				// Failed to read document. This is a failure.
				t.replaceExistingCounter.failed++
				t.reportFailure(test.NewFailure("Failed to read existing document '%s' (%s) in collection '%s' that should have been replaced: %v", key, ifMatchStatus, c.name, err))
				return "", maskAny(err)
			} else if !match {
				// The document does not match what we expect after the replace. This is a failure.
				t.replaceExistingCounter.failed++
				t.reportFailure(test.NewFailure("Failed to replace existing document '%s' (%s) in collection '%s': got 412 but has not been replaced", key, ifMatchStatus, c.name))
				return "", maskAny(fmt.Errorf("Failed to replace existing document '%s' (%s) in collection '%s': got 412 but has not been replaced", key, ifMatchStatus, c.name))
			} else {
				// Match found, document has been replaced after all
				update.Rev = rev
			}
		} else {
			// We got a 412 without asking for an explicit revision. This is a failure.
			t.replaceExistingCounter.failed++
			t.reportFailure(test.NewFailure("Failed to replace existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match", key, ifMatchStatus, c.name))
			return "", maskAny(fmt.Errorf("Failed to replace existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match", key, ifMatchStatus, c.name))
		}
	}
	// Update internal doc
	newDoc.rev = update.Rev
	c.existingDocs[key] = newDoc
	t.replaceExistingCounter.succeeded++
	t.log.Infof("Replacing existing document '%s' (%s) in '%s' (name -> '%s') succeeded", key, ifMatchStatus, c.name, newName)
	return update.Rev, nil
}

// replaceExistingDocumentWrongRevision replaces an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) replaceExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
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
	_, err := t.client.Put(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, newDoc, "", nil, []int{412}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.replaceExistingWrongRevisionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to replace existing document '%s' wrong revision in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.replaceExistingWrongRevisionCounter.succeeded++
	t.log.Infof("Replacing existing document '%s' wrong revision in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}

// replaceNonExistingDocument replaces a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) replaceNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
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
	if _, err := t.client.Put(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, newDoc, "", nil, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.replaceNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to replace non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.replaceNonExistingCounter.succeeded++
	t.log.Infof("Replacing non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}
