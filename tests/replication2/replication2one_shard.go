package replication2

import (
	"fmt"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type OneShardTest struct {
	DocColTest
	databaseName      string
	databaseNameSeq   int64
	isDatabaseCreated bool
}

func NewOneShardTest(log *logging.Logger, reportDir string, rep2config Replication2Config, config DocColConfig) test.TestScript {
	return &OneShardTest{
		DocColTest: DocColTest{
			Replication2Test: Replication2Test{
				TestName: "oneShardDbTest",
				Replication2TestContext: Replication2TestContext{
					Replication2Config: rep2config,
					Replication2TestHarness: Replication2TestHarness{
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
}

func (t *OneShardTest) generateCollectionName(seed int64) string {
	return "oneshard_docs_" + strconv.FormatInt(seed, 10)
}

func (t *OneShardTest) generateDatabaseName(seed int64) string {
	return "oneshard_db_" + strconv.FormatInt(seed, 10)
}

func (t *OneShardTest) testLoop() {
	t.active = true
	t.actions = 0
	defer func() { t.active = false }()

	var plan []int
	planIndex := 0
	for {
		// Should we stop
		if t.shouldStop() {
			return
		}
		if t.pauseRequested {
			t.paused = true
			time.Sleep(time.Second * 2)
			continue
		}
		t.paused = false
		t.actions++
		if plan == nil || planIndex >= len(plan) {
			plan = []int{0, 1, 2, 3, 4, 5} // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
			// create a database
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
			planIndex++

		case 1:
			// create a document collection
			if !t.docCollectionCreated && t.isDatabaseCreated {
				t.docCollectionName = t.generateCollectionName(t.collectionNameSeq)
				if err := t.createCollection(t.docCollectionName, false); err != nil {
					t.log.Errorf("Failed to create collection: %v", err)
				} else {
					t.docCollectionCreated = true
					t.actions++
				}
			}
			planIndex++

		case 2:
			// create documents
			t.DocColTest.createDocuments()
			planIndex++

		case 3:
			// read documents
			t.DocColTest.readDocuments()
			planIndex++

		case 4:
			t.DocColTest.updateDocuments()
			planIndex++

		case 5:
			// drop database
			if t.docCollectionCreated && t.numberOfExistingDocs >= t.MaxDocuments && t.existingDocuments[len(t.existingDocuments)-1].UpdateCounter > 10 {
				t.client.UseDatabase("_system")
				if err := t.dropDatabase(t.databaseName); err != nil {
					t.client.UseDatabase(t.databaseName)
					t.log.Errorf("Failed to drop database: %v", err)
				} else {
					t.isDatabaseCreated = false
					t.docCollectionCreated = false
					t.numberOfExistingDocs = 0
					t.existingDocuments = t.existingDocuments[:0]
					t.collectionNameSeq++
					t.databaseNameSeq++
					t.actions++
				}
			}
			planIndex++
		}
		time.Sleep(time.Second * 2)
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

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *OneShardTest) Start(cluster cluster.Cluster, listener test.TestListener) error {
	t.activeMutex.Lock()
	defer t.activeMutex.Unlock()

	if t.active {
		// No restart unless needed
		return nil
	}
	if err := t.setupLogger(cluster); err != nil {
		return maskAny(err)
	}

	t.cluster = cluster
	t.listener = listener
	t.client = util.NewArangoClient(t.log, cluster)

	t.active = true
	go t.testLoop()
	return nil
}
