package simple

import "fmt"

// isDocumentEqualTo reads an existing document and checks that it is equal to the given document.
// Returns: (isEqual,currentRevision,error)
func (t *simpleTest) isDocumentEqualTo(c *collection, key string, expected UserDocument) (bool, string, error) {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	var result UserDocument
	t.log.Infof("Checking existing document '%s' from '%s'...", key, c.name)
	resp, err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", c.name, key), nil, nil, &result, []int{200, 201, 202}, []int{400, 404, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.log.Errorf("Failed to read document '%s' from '%s': %v", key, c.name, err)
		return false, "", maskAny(err)
	}
	// Compare document against expected document
	if result.Equals(expected) {
		// Found an exact match
		return true, resp.Rev, nil
	}
	t.log.Infof("Document '%s' in '%s'  returned different values: got %q expected %q", key, c.name, result, expected)
	return false, resp.Rev, nil
}
