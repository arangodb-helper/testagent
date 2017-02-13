package simple

import (
	"fmt"
	"time"

	"github.com/arangodb/testAgent/service/test"
)

// readExistingDocument reads an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) readExistingDocument(collectionName string, key, rev string, updateRevision bool) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	var result UserDocument
	hdr, ifMatchStatus := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Reading existing document '%s' (%s) from '%s'...", key, ifMatchStatus, collectionName)
	if _, err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), nil, hdr, &result, []int{200, 201, 202}, []int{400, 404, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.readExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, collectionName, err))
		return maskAny(err)
	}
	// Compare document against expected document
	expected := t.existingDocs[key]
	if result.Value != expected.Value || result.Name != expected.Name || result.Odd != expected.Odd {
		// This is a failure
		t.readExistingCounter.failed++
		t.reportFailure(test.NewFailure("Read existing document '%s' (%s) returned different values '%s': got %q expected %q", key, ifMatchStatus, collectionName, result, expected))
		return maskAny(fmt.Errorf("Read returned invalid values"))
	}
	if updateRevision {
		// Store read document so we have the last revision
		t.existingDocs[key] = result
	}
	t.readExistingCounter.succeeded++
	t.log.Infof("Reading existing document '%s' (%s) from '%s' succeeded", key, ifMatchStatus, collectionName)
	return nil
}

// readExistingDocumentWrongRevision reads an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) readExistingDocumentWrongRevision(collectionName string, key, rev string, updateRevision bool) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	var result UserDocument
	hdr := ifMatchHeader(nil, rev)
	t.log.Infof("Reading existing document '%s' wrong revision from '%s'...", key, collectionName)
	if _, err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), nil, hdr, &result, []int{412}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.readExistingWrongRevisionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read existing document '%s' wrong revision in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.readExistingWrongRevisionCounter.succeeded++
	t.log.Infof("Reading existing document '%s' wrong revision from '%s' succeeded", key, collectionName)
	return nil
}

// readNonExistingDocument reads a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) readNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	var result UserDocument
	t.log.Infof("Reading non-existing document '%s' from '%s'...", key, collectionName)
	if _, err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), nil, nil, &result, []int{404}, []int{200, 201, 202, 400, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.readNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.readNonExistingCounter.succeeded++
	t.log.Infof("Reading non-existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}