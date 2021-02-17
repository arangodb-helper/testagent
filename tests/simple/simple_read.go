package simple

import (
	"fmt"

	"github.com/arangodb-helper/testagent/service/test"
)

// readExistingDocument reads an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) readExistingDocument(c *collection, key, rev string, updateRevision, skipExpectedValueCheck bool) (string, error) {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	var result UserDocument
	hdr, ifMatchStatus, _ := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Reading existing document '%s' (%s) from '%s'...", key, ifMatchStatus, c.name)
	if _, err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", c.name, key), nil, hdr, &result, []int{200, 201, 202}, []int{400, 404, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.readExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, c.name, err))
		return "", maskAny(err)
	}
	// Compare document against expected document
	if !skipExpectedValueCheck {
		expected := c.existingDocs[key]
		if result.Value != expected.Value || result.Name != expected.Name || result.Odd != expected.Odd {
			// This is a failure
			t.readExistingCounter.failed++
			t.reportFailure(test.NewFailure("Read existing document '%s' (%s) returned different values '%s': got %q expected %q", key, ifMatchStatus, c.name, result, expected))
			return "", maskAny(fmt.Errorf("Read returned invalid values"))
		}
	}
	if updateRevision {
		// Store read document so we have the last revision
		c.existingDocs[key] = result
	}
	t.readExistingCounter.succeeded++
	t.log.Infof("Reading existing document '%s' (%s) from '%s' succeeded", key, ifMatchStatus, c.name)
	return result.rev, nil
}

// readExistingDocumentWrongRevision reads an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) readExistingDocumentWrongRevision(collectionName string, key, rev string, updateRevision bool) error {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
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
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
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
