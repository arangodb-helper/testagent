package simple

import (
	"bytes"
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

// createImportDocument creates a #document based import file.
func (t *simpleTest) createImportDocument() ([]byte, []UserDocument) {
	buf := &bytes.Buffer{}
	docs := make([]UserDocument, 0, 10000)
	fmt.Fprintf(buf, `[ "_key", "value", "name", "odd" ]`)
	fmt.Fprintln(buf)
	for i := 0; i < 10000; i++ {
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
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("collection", c.name)
	q.Set("waitForSync", "true")

	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break;
		}

		importData, docs := t.createImportDocument()
		t.log.Infof("Importing %d documents ('%s' - '%s') into '%s'...",
			len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
		resp, err := t.client.Post("/_api/import", q, nil, importData, "application/x-www-form-urlencoded", nil,
			[]int{0, 1, 200, 201, 202, 503}, []int{400, 404, 409, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.importCounter.failed++
			t.reportFailure(test.NewFailure("Failed to import documents in collection '%s': %v", c.name, err[0]))
			return maskAny(err[0])
		}

		if resp[0].StatusCode >= 200 && resp[0].StatusCode <= 299 {
			for _, d := range docs {
				c.existingDocs[d.Key] = d
			}
			t.importCounter.succeeded++
			t.log.Infof("Importing (%d) %d documents ('%s' - '%s') into '%s' succeeded", i, len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
			return nil
		}

		t.log.Infof("Importing (%d) %d documents ('%s' - '%s') into '%s' got %d",
			i, len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name, resp[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	t.importCounter.failed++
	t.planCollectionDrop(c.name)
	t.reportFailure(test.NewFailure("Timed out while importing (%d) documents in collection '%s'", i, c.name))
	return maskAny(fmt.Errorf("Timed out while importing (%d) documents in collection '%s'", i, c.name))

}
