package replication2

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type DocColConfig struct {
	MaxDocuments int
	MaxEdges     int
	BatchSize    int
	DocumentSize int
}

type DocColTest struct {
	Replication2Test
	DocColConfig
	numberOfExistingDocs     int
	numberOfCreatedDocsTotal int64
	docCollectionCreated     bool
	docCollectionName        string
}

func NewDocColTest(log *logging.Logger, reportDir string, rep2config Replication2Config, config DocColConfig) test.TestScript {
	return &DocColTest{
		Replication2Test: Replication2Test{
			TestName: "documentCollectionTest",
			Replication2TestContext: Replication2TestContext{
				Replication2Config: rep2config,
				Replication2TestHarness: Replication2TestHarness{
					reportDir: reportDir,
					log:       log,
				},
				documentIdSeq:     0,
				collectionNameSeq: 0,
				existingDocSeeds:  make([]int64, 0, config.MaxDocuments),
			},
		},
		DocColConfig:             config,
		numberOfExistingDocs:     0,
		numberOfCreatedDocsTotal: 0,
		docCollectionCreated:     false,
	}
}

type BigDocument struct {
	TestDocument
	Value         int64  `json:"value"`
	Name          string `json:"name"`
	Odd           bool   `json:"odd"`
	UpdateCounter int    `json:"update_counter"`
	Payload       string `json:"payload"`
}

// Equals returns true when the value fields of `d` and `other` are the equal.
func (d BigDocument) Equals(other BigDocument) bool {
	return d.Value == other.Value && d.Name == other.Name && d.Odd == other.Odd && d.Payload == other.Payload
}

func (t *DocColTest) generateCollectionName(seed int64) string {
	return "replication2_docs_" + strconv.FormatInt(seed, 10)
}

func generateKeyFromSeed(seed int64) string {
	return strconv.FormatInt(seed, 10)
}

func NewBigDocument(seed int64, payloadSize int) BigDocument {
	randGen := rand.New(rand.NewSource(seed))
	payloadBytes := make([]byte, payloadSize)
	lowerBound := 32
	upperBound := 126
	for i := 0; i < payloadSize; i++ {
		payloadBytes[i] = byte(randGen.Int31n(int32(upperBound-lowerBound)) + int32(lowerBound))
	}
	var key string
	key = generateKeyFromSeed(seed)
	return BigDocument{
		TestDocument:  TestDocument{Key: key},
		Value:         seed,
		Name:          strconv.FormatInt(seed, 10),
		Odd:           seed%2 == 1,
		UpdateCounter: 0,
		Payload:       string(payloadBytes),
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
			plan = []int{0, 1, 2, 3, 4, 5} // Update when more tests are added
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
			if t.docCollectionCreated {
				for {
					if t.numberOfExistingDocs >= t.MaxDocuments-t.MaxEdges {
						break
					}
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
							t.existingDocSeeds = append(t.existingDocSeeds, seed)
							t.numberOfExistingDocs++
							t.numberOfCreatedDocsTotal++
						}
					}

					//FIXME: use bulk document creation here when createDocuments function is fixed

					// if err := t.createDocuments(thisBatchSize, t.documentIdSeq); err != nil {
					// 	t.log.Errorf("Failed to create documents: %#v", err)
					// } else {
					// 	t.numberOfExistingDocs += thisBatchSize
					// 	t.numberOfCreatedDocsTotal += int64(thisBatchSize)
					// }
					// t.documentIdSeq += int64(thisBatchSize)
					// t.actions++
				}
			}
			planIndex++

		case 2:
			// read documents
			if t.docCollectionCreated {
				for _, seed := range t.existingDocSeeds {
					expectedDocument := NewBigDocument(seed, t.DocumentSize)
					if err := t.readExistingDocument(t.docCollectionName, expectedDocument, false); err != nil {
						t.log.Errorf("Failed to read document: %v", err)
						t.readExistingCounter.failed++
					} else {
						t.actions++
						t.readExistingCounter.succeeded++
					}
				}
			}
			planIndex++

		case 3:
			// update documents
			// if t.docCollectionCreated {
			// 	for _, seed := range t.existingDocSeeds {
			// 		if err := t.readExistingDocument(t.docCollectionName, seed, false); err != nil {
			// 			t.log.Errorf("Failed to read document: %v", err)
			// 			t.readExistingCounter.failed++
			// 		} else {
			// 			t.actions++
			// 			t.readExistingCounter.succeeded++
			// 		}
			// 	}
			// }
			planIndex++

		case 4:
			// read documents again after update
			planIndex++

		case 5:
			// drop collections
			if t.docCollectionCreated && t.numberOfExistingDocs >= t.MaxDocuments {
				if err := t.dropCollection(t.docCollectionName); err != nil {
					t.log.Errorf("Failed to drop collection: %v", err)
				} else {
					t.docCollectionCreated = false
					t.numberOfExistingDocs = 0
					t.existingDocSeeds = t.existingDocSeeds[:0]
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
