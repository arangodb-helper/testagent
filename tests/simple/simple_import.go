package simple

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/arangodb/testAgent/service/test"
)

const (
	ndocs = 10000
)

// createImportDocument creates a #document based import file.
func (t *simpleTest) createImportDocument() ([]byte, []UserDocument) {
	buf := &bytes.Buffer{}
	docs := make([]UserDocument, 0, ndocs)
	fmt.Fprintf(buf, `[ "_key", "value", "name", "odd" ]`)
	fmt.Fprintln(buf)
	for i := 0; i < ndocs; i++ {
		key := fmt.Sprintf("docimp%05d", i)
		userDoc := UserDocument{
			Key:   key,
			Value: i,
			Name:  fmt.Sprintf("Imported %d", i),
			Odd:   i%2 == 0,
		}
		docs = append(docs, userDoc)
		fmt.Fprintf(buf, `[ "%s", %d, "%s", %v ]`, userDoc.Key, userDoc.Value, userDoc.Name, userDoc.Odd)
		fmt.Fprintln(buf)
	}
	return buf.Bytes(), docs
}

// importDocuments imports a bulk set of documents.
// The operation is expected to succeed.
func (t *simpleTest) importDocuments(c *collection) error {
	operationTimeout, retryTimeout := t.OperationTimeout*4, t.RetryTimeout*4
	q := url.Values{}
	q.Set("collection", c.name)
	q.Set("waitForSync", "true")
	q.Set("details", "true")
	importData, docs := t.createImportDocument()
	t.log.Infof("Importing %d documents ('%s' - '%s') into '%s'...", len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
	var result interface{}
	if _, err := t.client.Post("/_api/import", q, nil, importData, "application/x-www-form-urlencoded", &result, []int{200, 201, 202}, []int{400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.importCounter.failed++
		t.reportFailure(test.NewFailure("Failed to import documents in collection '%s': %v", c.name, err))
		return maskAny(err)
	}

	switch v := result.(type) {
	case map[string]interface{}:
		if created, ok := v["created"]; ok {
			if cint, ok := created.(float64); ok {
				if cint != ndocs {
					// We do not create a failure here, since some chaos can
					// always prevent an import from going through. However,
					// this incident will be logged and the half-imported
					// collection will by left in the database for later
					// inspecting.
					if details, ok := v["details"]; ok {
						t.importCounter.failed++
						return maskAny(fmt.Errorf("Failed to import documents in collection '%s': incomplete import, details: %v", c.name, details))
					} else { // details missing although error
						return maskAny(fmt.Errorf("Failed to import documents in collection '%s': incomplete import, no details", c.name))
					}
				} // no import off
			}
		}
	default:
		t.importCounter.failed++
		t.reportFailure(
			test.NewFailure("Failed to import documents in collection '%s': unexpected result %v", c.name, result))
		return maskAny(fmt.Errorf("Failed to import documents in collection '%s': unexpected result %v", c.name, result))
	}

	for _, d := range docs {
		c.existingDocs[d.Key] = d
	}
	t.importCounter.succeeded++
	t.log.Infof("Importing %d documents ('%s' - '%s') into '%s' succeeded", len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
	return nil
}
