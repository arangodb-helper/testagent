package simple

import (
	"bytes"
	"fmt"
	"net/url"
	"time"
	"github.com/arangodb-helper/testagent/service/test"
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
	operationTimeout := t.OperationTimeout * 4
	testTimeout := time.Now().Add(t.OperationTimeout * 20)

	q := url.Values{}
	q.Set("collection", c.name)
	q.Set("waitForSync", "true")

	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		importData, docs := t.createImportDocument()
		var result interface{}
		t.log.Infof("Importing %d documents ('%s' - '%s') into '%s'...",
			len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
		resp, err := t.client.Post("/_api/import", q, nil, importData, "application/x-www-form-urlencoded", &result,
			[]int{0, 1, 200, 201, 202, 503}, []int{400, 404, 409, 307}, operationTimeout, 1)
		t.log.Infof("... got http %d - arangodb %d via %s",
			resp[0].StatusCode, resp[0].Error_.ErrorNum, resp[0].CoordinatorURL)

		if err[0] != nil {
			// This is a failure
			t.importCounter.failed++
			t.reportFailure(test.NewFailure("Failed to import documents in collection '%s': %v", c.name, err[0]))
			return maskAny(err[0])
		}

		if resp[0].StatusCode >= 200 && resp[0].StatusCode <= 299 {
			switch v := result.(type) {
			case map[string]interface{}:
				if created, ok := v["created"]; ok {
					if cint, ok := created.(float64); ok {
						if cint != ndocs {
							if details, ok := v["details"]; ok {
								t.importCounter.failed++
								t.reportFailure(
									test.NewFailure(
										"Failed to import documents in collection '%s': incomplete import of only %d documents, details: %v",
										c.name, cint, details))
								return maskAny(fmt.Errorf("Failed to import documents in collection '%s': incomplete import"))
							} else { // details missing although error
								t.reportFailure(test.NewFailure(
									"Failed to import documents in collection '%s': incomplete import of only %d documents no details(!!!)",
									c.name, cint))
								return maskAny(fmt.Errorf("Failed to import documents in collection '%s': incomplete import"))
							}
						} // no import off
					} else { // cint not int64 convertible
						t.reportFailure(test.NewFailure(
							"Failed to import documents in collection '%s': invalid response on import, cannot convert import count %T %+v",
							c.name, created, created))
						return maskAny(fmt.Errorf(
							"Failed to import documents in collection '%s': invalid response on import, cannot convert import count %T %+v",
							c.name, created, created))
					}
				} else { // no created key in result
					t.reportFailure(test.NewFailure(
						"Failed to import documents in collection '%s': invalid response on import, no 'created' key in result",
						c.name))
					return maskAny(fmt.Errorf(
						"Failed to import documents in collection '%s': invalid response on import, , no 'created' key in result",
						c.name))
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
			t.log.Infof("Importing (%d) %d documents ('%s' - '%s') into '%s' succeeded", i, len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
			return nil
		}

		time.Sleep(backoff)
		if backoff < time.Second*5 {
			backoff += backoff
		}

	}

	t.importCounter.failed++
	t.planCollectionDrop(c.name)
	t.reportFailure(test.NewFailure("Timed out while importing (%d) documents in collection '%s'", i, c.name))
	return maskAny(fmt.Errorf("Timed out while importing (%d) documents in collection '%s'", i, c.name))

}
