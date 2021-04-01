package simple

import (
	"fmt"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

// readExistingDocument reads an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) readExistingDocument(
	c *collection, key, rev string, updateRevision, skipExpectedValueCheck bool) (string, error) {

	operationTimeout := t.OperationTimeout/5
	testTimeout := time.Now().Add(operationTimeout)

	i := 0
	url := fmt.Sprintf("/_api/document/%s/%s", c.name, key)

	var result UserDocument
	hdr, ifMatchStatus, _ := createRandomIfMatchHeader(nil, rev)

	for {

		if time.Now().After(testTimeout) {
			break;
		}
		i++;

		t.log.Infof("Reading existing document '%s' (%s) from '%s'...", key, ifMatchStatus, c.name)
		resp, err := t.client.Get(
			url, nil, hdr, &result, []int{0, 1, 200, 503}, []int{400, 404, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.readExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(
					"Failed to read existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, c.name, err[0]))
			return "", maskAny(err[0])
		} else {
			if resp[0].StatusCode == 200 {
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
		}
	}

	t.readExistingCounter.failed++
	t.reportFailure(
		test.NewFailure(
			"Timed out (%d) reading existing document '%s' from %s", i, key, c.name))
	return "", maskAny(fmt.Errorf("Timed out (%d) reading existing document '%s' from %s", i, key, c.name))

}

// readExistingDocumentWrongRevision reads an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) readExistingDocumentWrongRevision(
	collectionName string, key, rev string, updateRevision bool) error {

	operationTimeout := t.OperationTimeout / 5
	testTimeout := time.Now().Add(operationTimeout)

	i := 0
	backoff := time.Millisecond * 250
	url := fmt.Sprintf("/_api/document/%s/%s", collectionName, key)
	var result UserDocument
	hdr := ifMatchHeader(nil, rev)

	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++

		t.log.Infof("Reading (%d) existing document '%s' wrong revision from '%s'...", i, key, collectionName)
		resp , err := t.client.Get(
			url, nil, hdr, &result, []int{0, 1, 412, 503}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.readExistingWrongRevisionCounter.failed++
			t.reportFailure(
				test.NewFailure(
					"Failed to read existing document '%s' wrong revision in collection '%s': %v",
					key, collectionName, err[0]))
			return maskAny(err[0])
		} else if resp[0].StatusCode == 412 {
			t.readExistingWrongRevisionCounter.succeeded++
			t.log.Infof("Reading existing document '%s' wrong revision from '%s' succeeded", key, collectionName)
			return nil
		}

		time.Sleep(backoff)
		backoff += backoff

	}

	t.readExistingWrongRevisionCounter.failed++
	t.reportFailure(test.NewFailure(
		"Timed out (%d) while reading existing document '%s' wrong revision in collection '%s'",
		i, key, collectionName))
	return maskAny(fmt.Errorf(
		"Timed out (%d) while reading existing document '%s' wrong revision in collection '%s'",
		i, key, collectionName))

}

// readNonExistingDocument reads a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) readNonExistingDocument(collectionName string, key string) error {

	operationTimeout := t.OperationTimeout / 5
	testTimeout := time.Now().Add(operationTimeout)

	i := 0
	backoff := time.Millisecond * 250
	url := fmt.Sprintf("/_api/document/%s/%s", collectionName, key)
	var result UserDocument

	for {

		if time.Now().After(testTimeout) {
			break
		}
		i++

		t.log.Infof("Reading (%d) non-existing document '%s' from '%s'...", i, key, collectionName)
		resp, err := t.client.Get(
			url, nil, nil, &result,[]int{0, 1, 404, 503}, []int{200, 201, 202, 400, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.readNonExistingCounter.failed++
			t.reportFailure(test.NewFailure(
				"Failed to read non-existing document '%s' in collection '%s': %v", key, collectionName, err[0]))
			return maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			t.readNonExistingCounter.succeeded++
			t.log.Infof("Reading non-existing document '%s' from '%s' succeeded", key, collectionName)
			return nil
		}

		time.Sleep(backoff)
		backoff += backoff

	}

	t.readNonExistingCounter.failed++
	t.reportFailure(test.NewFailure(
		"Timed out while reading non-existing document '%s' in collection '%s'", i, key, collectionName))
	return maskAny(fmt.Errorf(
		"Timed out while reading non-existing document '%s' in collection '%s'", i, key, collectionName))
}
