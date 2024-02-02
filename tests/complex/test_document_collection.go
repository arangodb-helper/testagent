package complex

import (
	"fmt"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type DocColConfig struct {
	MaxDocuments int
	MaxUpdates   int
	BatchSize    int
	DocumentSize int
}

type DocColTest struct {
	ComplextTest
	DocColConfig
	numberOfExistingDocs     int
	numberOfCreatedDocsTotal int64
	docCollectionCreated     bool
	docCollectionName        string
	readOffset               int
	updateOffset             int
}

func NewDocColTest(log *logging.Logger, reportDir string, rep2config ComplextTestConfig, config DocColConfig) test.TestScript {
	return &DocColTest{
		ComplextTest: ComplextTest{
			TestName: "documentCollectionTest",
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
		readOffset:               0,
	}
}

// Equals returns true when the value fields of `d` and `other` are the equal.
func (d BigDocument) Equals(other BigDocument) bool {
	return d.Value == other.Value && d.Name == other.Name && d.Odd == other.Odd && d.Payload == other.Payload && d.UpdateCounter == other.UpdateCounter
}

func (t *DocColTest) generateCollectionName(seed int64) string {
	return "documents_" + strconv.FormatInt(seed, 10)
}

func generateKeyFromSeed(seed int64) string {
	return strconv.FormatInt(seed, 10)
}

func (t *DocColTest) createDocuments() {
	if t.docCollectionCreated && t.numberOfExistingDocs < t.MaxDocuments {
		var thisBatchSize int
		if t.BatchSize <= t.MaxDocuments-t.numberOfExistingDocs {
			thisBatchSize = t.BatchSize
		} else {
			thisBatchSize = t.MaxDocuments - t.numberOfExistingDocs
		}

		for i := 0; i < thisBatchSize; i++ {
			seed := t.documentIdSeq
			t.documentIdSeq++
			document := NewBigDocument(seed, t.DocumentSize)
			if err := t.insertDocument(t.docCollectionName, document); err != nil {
				t.log.Errorf("Failed to create document with key '%s' in collection '%s': %v",
					document.Key, t.docCollectionName, err)
			} else {
				t.actions++
				t.existingDocuments = append(t.existingDocuments, document.TestDocument)
				t.numberOfExistingDocs++
				t.numberOfCreatedDocsTotal++
			}
		}
	}
}

func (t *DocColTest) readDocuments() {
	if t.docCollectionCreated && t.numberOfExistingDocs >= t.BatchSize {
		var upperBound int = 0
		var lowerBound int = t.readOffset
		if t.numberOfExistingDocs-t.readOffset < t.BatchSize {
			upperBound = t.numberOfExistingDocs
		} else {
			upperBound = t.readOffset + t.BatchSize
		}

		for _, testDoc := range t.existingDocuments[lowerBound:upperBound] {
			expectedDocument := NewBigDocumentFromTestDocument(testDoc, t.DocumentSize)
			if err := t.readExistingDocument(t.docCollectionName, expectedDocument, false); err != nil {
				t.log.Errorf("Failed to read document: %v", err)
			} else {
				t.actions++
			}
		}
		if upperBound == t.numberOfExistingDocs {
			t.readOffset = 0
		} else {
			t.readOffset = upperBound
		}
	}
}

func (t *DocColTest) updateDocuments() {
	if t.docCollectionCreated && t.numberOfExistingDocs >= t.BatchSize {
		if t.updateOffset == t.numberOfExistingDocs {
			t.updateOffset = 0
		}
		var upperBound int = 0
		var lowerBound int = t.updateOffset
		if t.numberOfExistingDocs-t.updateOffset < t.BatchSize {
			upperBound = t.numberOfExistingDocs
		} else {
			upperBound = t.updateOffset + t.BatchSize
		}
		for i := lowerBound; i < upperBound; i++ {
			oldDoc := t.existingDocuments[i]
			if newDoc, err := t.updateExistingDocument(t.docCollectionName, oldDoc); err != nil {
				t.log.Errorf("Failed to update document: %v", err)
			} else {
				t.existingDocuments[i] = *newDoc
				t.actions++
			}
		}
		t.updateOffset = upperBound
	}
}

func (t *DocColTest) testLoop() {
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
			plan = []int{0, 1, 2, 3, 4} // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
			// create a document collection
			if !t.docCollectionCreated {
				t.docCollectionName = t.generateCollectionName(t.collectionNameSeq)
				if err := t.createCollection(t.docCollectionName, false); err != nil {
					t.log.Errorf("Failed to create collection: %v", err)
				} else {
					t.docCollectionCreated = true
					t.actions++
				}
			}
			planIndex++

		case 1:
			// create documents
			t.createDocuments()
			planIndex++

		case 2:
			// read documents
			t.readDocuments()
			planIndex++

		case 3:
			// update documents
			t.updateDocuments()
			planIndex++

		case 4:
			// drop collections
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
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}

// Status returns the current status of the test
func (t *DocColTest) Status() test.TestStatus {
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

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *DocColTest) Start(cluster cluster.Cluster, listener test.TestListener) error {
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
