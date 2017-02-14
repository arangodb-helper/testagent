package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb/testAgent/service/test"
)

// updateExistingDocument updates an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) updateExistingDocument(c *collection, key, rev string) (string, error) {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr, ifMatchStatus := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Updating existing document '%s' (%s) in '%s' (name -> '%s')...", key, ifMatchStatus, c.name, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	doc := c.existingDocs[key]
	update, err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", c.name, key), q, hdr, delta, "", nil, []int{200, 201, 202}, []int{400, 404, 412, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.updateExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, c.name, err))
		return "", maskAny(err)
	}
	// Update internal doc
	doc.Name = newName
	doc.rev = update.Rev
	c.existingDocs[key] = doc
	t.updateExistingCounter.succeeded++
	t.log.Infof("Updating existing document '%s' (%s) in '%s' (name -> '%s') succeeded", key, ifMatchStatus, c.name, newName)
	return update.Rev, nil
}

// updateExistingDocumentWrongRevision updates an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) updateExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr := ifMatchHeader(nil, rev)
	t.log.Infof("Updating existing document '%s' wrong revision in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	_, err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, delta, "", nil, []int{412}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.updateExistingWrongRevisionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update existing document '%s' wrong revision in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.updateExistingWrongRevisionCounter.succeeded++
	t.log.Infof("Updating existing document '%s' wrong revision in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}

// updateNonExistingDocument updates a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) updateNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated non-existing name %s", time.Now())
	t.log.Infof("Updating non-existing document '%s' in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	if _, err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, delta, "", nil, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.updateNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.updateNonExistingCounter.succeeded++
	t.log.Infof("Updating non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}
