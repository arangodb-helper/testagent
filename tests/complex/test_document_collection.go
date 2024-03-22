package complex

import (
	"fmt"
	"strconv"

	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
)

type RegularDocColTest struct {
	DocColTest
}

func NewRegularDocColTest(log *logging.Logger, reportDir string, complexTestCfg ComplextTestConfig, config DocColConfig) test.TestScript {
	docColTest := &RegularDocColTest{
		DocColTest: DocColTest{
			ComplextTest: ComplextTest{
				TestName: "documentCollectionTest",
				ComplextTestContext: ComplextTestContext{
					ComplextTestConfig: complexTestCfg,
					ComplextTestHarness: ComplextTestHarness{
						reportDir: reportDir,
						log:       log,
					},
					documentIdSeq:     0,
					collectionNameSeq: 0,
					existingDocuments: make([]TestDocument, 0, config.MaxDocuments),
				},
			},
			DocColConfig:             config,
			numberOfExistingDocs:     0,
			numberOfCreatedDocsTotal: 0,
			docCollectionCreated:     false,
			readOffset:               0,
		}}
	docColTest.DocColTestImpl = docColTest
	docColTest.ComplexTestImpl = docColTest
	return docColTest
}

func (t *RegularDocColTest) generateCollectionName(seed int64) string {
	return "documents_" + strconv.FormatInt(seed, 10)
}

func (t *RegularDocColTest) createTestDatabase() {
	//we do not need to create a DB, since we use _system DB for this test
}

func (t *RegularDocColTest) dropTestDatabase() {
	//we do not need to drop a DB, since we use _system DB for this test
}

func (t *RegularDocColTest) createTestCollection() {
	if !t.docCollectionCreated {
		t.docCollectionName = t.generateCollectionName(t.collectionNameSeq)
		if err := t.createCollection(t.docCollectionName, false); err != nil {
			t.log.Errorf("Failed to create collection: %v", err)
		} else {
			t.docCollectionCreated = true
			t.actions++
		}
	}
}

func (t *RegularDocColTest) dropTestCollection() {
	if t.docCollectionCreated && t.numberOfExistingDocs >= t.MaxDocuments && t.existingDocuments[len(t.existingDocuments)-1].UpdateCounter > t.MaxUpdates {
		if err := t.dropCollection(t.docCollectionName); err != nil {
			t.log.Errorf("Failed to drop collection: %v", err)
		} else {
			t.docCollectionCreated = false
			t.numberOfExistingDocs = 0
			t.existingDocuments = t.existingDocuments[:0]
			t.updateOffset = 0
			t.readOffset = 0
			t.readOffset = 0
			t.collectionNameSeq++
			t.actions++
		}
	}
}

// Status returns the current status of the test
func (t *RegularDocColTest) Status() test.TestStatus {
	cc := func(name string, c counter) test.Counter {
		return test.Counter{
			Name:      name,
			Succeeded: c.succeeded,
			Failed:    c.failed,
		}
	}

	status := test.TestStatus{
		Active:   t.active && !t.paused,
		Pausing:  t.pauseRequested,
		Failures: t.failures,
		Actions:  t.actions,
		Counters: []test.Counter{
			cc("#collections created", t.createCollectionCounter),
			cc("#collections dropped", t.dropCollectionCounter),
			cc("#document batches created", t.bulkCreateCounter),
			cc("#single documents created", t.singleDocCreateCounter),
			cc("#documents read", t.readExistingCounter),
			cc("#documents updated", t.updateExistingCounter),
			cc("#documents replaced", t.replaceExistingCounter),
		},
	}

	status.Messages = append(status.Messages,
		fmt.Sprintf("Number of documents in the database: %d", t.numberOfExistingDocs),
		fmt.Sprintf("Number of shards in the collection: %d", t.NumberOfShards),
	)

	return status
}
