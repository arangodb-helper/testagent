package complex

import (
	"fmt"
	"strconv"

	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
)

type OneShardTest struct {
	DocColTest
	databaseName      string
	databaseNameSeq   int64
	isDatabaseCreated bool
}

func NewOneShardTest(log *logging.Logger, reportDir string, rep2config ComplextTestConfig, config DocColConfig) test.TestScript {
	oneShardTest := &OneShardTest{
		DocColTest: DocColTest{
			ComplextTest: ComplextTest{
				TestName: "oneShardDbTest",
				ComplextTestContext: ComplextTestContext{
					ComplextTestConfig: rep2config,
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
		},
		databaseNameSeq:   0,
		isDatabaseCreated: false,
	}
	oneShardTest.DocColTestImpl = oneShardTest
	oneShardTest.ComplexTestImpl = oneShardTest
	return oneShardTest
}

func (t *OneShardTest) generateCollectionName(seed int64) string {
	return "oneshard_docs_" + strconv.FormatInt(seed, 10)
}

func (t *OneShardTest) generateDatabaseName(seed int64) string {
	return "oneshard_db_" + strconv.FormatInt(seed, 10)
}

func (t *OneShardTest) createTestDatabase() {
	if !t.isDatabaseCreated {
		t.databaseName = t.generateDatabaseName(t.databaseNameSeq)
		if err := t.createOneShardDatabase(t.databaseName); err != nil {
			t.log.Errorf("Failed to create database: %v", err)
		} else {
			t.isDatabaseCreated = true
			t.client.UseDatabase(t.databaseName)
			t.actions++
		}
	}
}

func (t *OneShardTest) createTestCollection() {
	if !t.docCollectionCreated && t.isDatabaseCreated {
		t.docCollectionName = t.generateCollectionName(t.collectionNameSeq)
		if err := t.createCollection(t.docCollectionName, false); err != nil {
			t.log.Errorf("Failed to create collection: %v", err)
		} else {
			t.docCollectionCreated = true
			t.actions++
		}
	}
}

func (t *OneShardTest) dropTestCollection() {
	//we do not need to drop collection becasue we will drop the database altogether
}

func (t *OneShardTest) dropTestDatabase() {
	if t.docCollectionCreated && t.numberOfExistingDocs >= t.MaxDocuments && t.existingDocuments[len(t.existingDocuments)-1].UpdateCounter > t.MaxUpdates {
		t.client.UseDatabase("_system")
		if err := t.dropDatabase(t.databaseName); err != nil {
			t.client.UseDatabase(t.databaseName)
			t.log.Errorf("Failed to drop database: %v", err)
		} else {
			t.isDatabaseCreated = false
			t.docCollectionCreated = false
			t.numberOfExistingDocs = 0
			t.existingDocuments = t.existingDocuments[:0]
			t.readOffset = 0
			t.updateOffset = 0
			t.collectionNameSeq++
			t.databaseNameSeq++
			t.actions++
		}
	}
}

// Status returns the current status of the test
func (t *OneShardTest) Status() test.TestStatus {
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
			cc("#databases created", t.createDatabaseCounter),
			cc("#databases dropped", t.dropDatabaseCounter),
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
	)

	return status
}
