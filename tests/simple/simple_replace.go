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
func (t *simpleTest) replaceExistingDocument(collectionName string, key, rev string) (string, error) {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr, ifMatchStatus := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Replacing existing document '%s' (%s) in '%s' (name -> '%s')...", key, ifMatchStatus, collectionName, newName)
	newDoc := UserDocument{
		Key:   key,
		Name:  fmt.Sprintf("Replaced named %s", key),
		Value: rand.Int(),
		Odd:   rand.Int()%2 == 0,
	}
	update, err := t.client.Put(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, newDoc, "", nil, []int{200, 201, 202}, []int{400, 404, 412, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.replaceExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to replace existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, collectionName, err))
		return "", maskAny(err)
	}
	// Update internal doc
	newDoc.rev = update.Rev
	t.existingDocs[key] = newDoc
	t.replaceExistingCounter.succeeded++
	t.log.Infof("Replacing existing document '%s' (%s) in '%s' (name -> '%s') succeeded", key, ifMatchStatus, collectionName, newName)
	return update.Rev, nil
}

// replaceExistingDocumentWrongRevision replaces an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) replaceExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
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
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
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
