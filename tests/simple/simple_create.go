package simple

import (
  "fmt"
  "net/url"

  "github.com/arangodb-helper/testagent/service/test"
)

// createDocument creates a new document.
// The operation is expected to succeed.
func (t *simpleTest) createDocument(c *collection, document interface{}, key string) (string, error) {
  operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
  q := url.Values{}
  q.Set("waitForSync", "true")
  t.log.Infof("Creating document '%s' in '%s'...", key, c.name)
	attempts := 4
	
	for i := 0; i < attempts; i++ {
		resp, err := t.client.Post(
			fmt.Sprintf("/_api/document/%s", c.name), q, nil, document, "", nil, []int{200, 201, 202, 409},
			[]int{400, 404, 307}, operationTimeout, 1)
		if err[0] != nil {
			t.lastRequestErr = false
			t.createCounter.failed++
			if (i == attempts-1) {
				t.reportFailure(test.NewFailure("Failed to create document with key '%s' in collection '%s': %v", key, c.name, err))
				return "", maskAny(err)
			}
		} else if resp.StatusCode == 409 {
			t.lastRequestErr = false
			// Duplicate key, check if this is correct
			if rev, err := t.readExistingDocument(c, key, "", true, true); err != nil {
				// Document with reported duplicate key cannot be read, so 409 status is a failure
				t.createCounter.failed++
				t.reportFailure(test.NewFailure("Failed to create document with key '%s' in collection '%s': got status 409, but read of document failed: %v", key, c.name, err))
				return "", maskAny(err)
			} else {
				// Use the revision we just read to avoid future failures
				resp.Rev = rev
			}
		} else {
			t.lastRequestErr = true
		}
	}
  t.createCounter.succeeded++
  t.log.Infof("Creating document '%s' in '%s' succeeded", key, c.name)
  return resp.Rev, nil
}
