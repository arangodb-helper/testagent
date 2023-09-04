package replication2

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type Replication2Config struct {
	MaxDocuments      int
	MaxEdges          int
	BatchSize         int
	DocumentSize      int
	NumberOfShards    int
	ReplicationFactor int
	OperationTimeout  time.Duration
	RetryTimeout      time.Duration
}

type replication2Test struct {
	Replication2Config
	activeMutex               sync.Mutex
	logPath                   string
	reportDir                 string
	log                       *logging.Logger
	cluster                   cluster.Cluster
	listener                  test.TestListener
	stop                      chan struct{}
	active                    bool
	pauseRequested            bool
	paused                    bool
	lastRequestErr            bool
	client                    util.ArangoClientInterface
	failures                  int
	actions                   int
	docCollectionName         string
	edgeCollectionName        string
	graphName                 string
	docCollectionCreated      bool
	edgeCollectionCreated     bool
	graphCreated              bool
	numberOfDocsThisDb        int
	numberOfCreatedDocsTotal  int64
	documentIdSeq             int64
	collectionNameSeq         int64
	existingDocSeeds          []int64
	createCollectionCounter   counter
	createGraphCounter        counter
	dropCollectionCounter     counter
	dropGraphCounter          counter
	singleDocCreateCounter    counter
	edgeDocumentCreateCounter counter
	bulkCreateCounter         counter
	readExistingCounter       counter
	updateExistingCounter     counter
}

type counter struct {
	succeeded int
	failed    int
}

// NewReplication2Test creates a replication2 test
func NewReplication2Test(log *logging.Logger, reportDir string, config Replication2Config) test.TestScript {
	return &replication2Test{
		Replication2Config:   config,
		reportDir:            reportDir,
		log:                  log,
		docCollectionCreated: false,
		documentIdSeq:        0,
		collectionNameSeq:    0,
		existingDocSeeds:     make([]int64, 0, config.MaxDocuments),
	}
}

func generateCollectionName(seed int64) string {
	return "replication2_docs_" + strconv.FormatInt(seed, 10)
}

func generateEdgeCollectionName(seed int64) string {
	return "replication2_edges_" + strconv.FormatInt(seed, 10)
}

func generateGraphName(seed int64) string {
	return "simple_named_graph_" + strconv.FormatInt(seed, 10)
}

// Name returns the name of the script
func (t *replication2Test) Name() string {
	return "replication2"
}

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *replication2Test) Start(cluster cluster.Cluster, listener test.TestListener) error {
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

// Stop any running test. This should not return until tests are actually stopped.
func (t *replication2Test) Stop() error {
	t.activeMutex.Lock()
	defer t.activeMutex.Unlock()

	if !t.active {
		// No active, nothing to stop
		return nil
	}

	stop := make(chan struct{})
	t.stop = stop
	<-stop
	return nil
}

// Interrupt the tests, but be prepared to continue.
func (t *replication2Test) Pause() error {
	t.pauseRequested = true
	return nil
}

// Resume running the tests, where Pause interrupted it.
func (t *replication2Test) Resume() error {
	t.pauseRequested = false
	return nil
}

// Status returns the current status of the test
func (t *replication2Test) Status() test.TestStatus {
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
		fmt.Sprintf("Number of documents in the database: %d", t.numberOfDocsThisDb),
		fmt.Sprintf("Number of shards in the collection: %d", t.NumberOfShards),
	)

	return status
}

// CollectLogs copies all logging info to the given writer.
func (t *replication2Test) CollectLogs(w io.Writer) error {
	if logPath := t.logPath; logPath == "" {
		// Nothing to log yet
		return nil
	} else {
		rd, err := os.Open(logPath)
		if err != nil {
			return maskAny(err)
		}
		defer rd.Close()
		if _, err := io.Copy(w, rd); err != nil {
			return maskAny(err)
		}
		return nil
	}
}

// setupLogger creates a new logger that is backed by stderr AND a file.
func (t *replication2Test) setupLogger(cluster cluster.Cluster) error {
	t.logPath = filepath.Join(t.reportDir, fmt.Sprintf("replication2-%s.log", cluster.ID()))
	logFile, err := os.Create(t.logPath)
	if err != nil {
		return maskAny(err)
	}
	fileBackend := logging.NewLogBackend(logFile, "", stdlog.LstdFlags)
	log := logging.MustGetLogger("simple")
	log.SetBackend(logging.MultiLogger(fileBackend, logging.NewLogBackend(os.Stderr, "", stdlog.LstdFlags)))
	t.log = log
	return nil
}

func (t *replication2Test) shouldStop() bool {
	// Should we stop?
	if stop := t.stop; stop != nil {
		stop <- struct{}{}
		return true
	}
	return false
}

func (t *replication2Test) reportFailure(f test.Failure) {
	t.failures++
	t.listener.ReportFailure(f)
}

func (t *replication2Test) testLoop() {
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
			plan = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
			// create a document collection
			if !t.docCollectionCreated {
				t.docCollectionName = generateCollectionName(t.collectionNameSeq)
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
					if t.numberOfDocsThisDb >= t.MaxDocuments-t.MaxEdges {
						break
					}
					var thisBatchSize int
					if t.BatchSize <= t.MaxDocuments-t.numberOfDocsThisDb {
						thisBatchSize = t.BatchSize
					} else {
						thisBatchSize = t.MaxDocuments - t.numberOfDocsThisDb
					}

					for i := 0; i < thisBatchSize; i++ {
						t.createDocument()
						t.actions++
						t.numberOfDocsThisDb++
						t.numberOfCreatedDocsTotal++
					}

					//FIXME: use bulk document creation here when createDocuments function is fixed

					// if err := t.createDocuments(thisBatchSize, t.documentIdSeq); err != nil {
					// 	t.log.Errorf("Failed to create documents: %#v", err)
					// } else {
					// 	t.numberOfDocsThisDb += thisBatchSize
					// 	t.numberOfCreatedDocsTotal += int64(thisBatchSize)
					// }
					// t.documentIdSeq += int64(thisBatchSize)
					// t.actions++
				}
			}
			planIndex++

		case 2:
			// create edges
			if t.edgeCollectionCreated {
				for {
					if t.numberOfDocsThisDb >= t.MaxEdges {
						break
					}
					var thisBatchSize int
					if t.BatchSize <= t.MaxEdges-t.numberOfDocsThisDb {
						thisBatchSize = t.BatchSize
					} else {
						thisBatchSize = t.MaxEdges - t.numberOfDocsThisDb
					}

					for i := 0; i < thisBatchSize; i++ {
						t.createEdge()
						t.actions++
						t.numberOfDocsThisDb++
						t.numberOfCreatedDocsTotal++
					}
				}
			}
			planIndex++

		case 3:
			// read documents
			if t.docCollectionCreated {
				for _, seed := range t.existingDocSeeds {
					if err := t.readExistingDocument(seed, false); err != nil {
						t.log.Errorf("Failed to read document: %v", err)
						t.readExistingCounter.failed++
					} else {
						t.actions++
						t.readExistingCounter.succeeded++
					}
				}
			}
			planIndex++

		case 4:
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

		case 5:
			// read documents again after update
			planIndex++

		case 6:
			//create an edge collection
			if t.docCollectionCreated && !t.edgeCollectionCreated {
				t.edgeCollectionName = generateEdgeCollectionName(t.collectionNameSeq)
				if err := t.createCollection(t.edgeCollectionName, true); err != nil {
					t.log.Errorf("Failed to create collection: %v", err)
				} else {
					t.edgeCollectionCreated = true
					t.actions++
				}
			}
			planIndex++

		case 7:
			//create a named graph
			if t.docCollectionCreated && t.edgeCollectionCreated {
				t.graphName = generateGraphName(t.collectionNameSeq)
				if err := t.createGraph(t.graphName, t.edgeCollectionName, []string{t.docCollectionName}, []string{t.docCollectionName},
					nil, false, false, "", nil, 0, 0, 0); err != nil {
					t.log.Errorf("Failed to create graph: %v", err)
				} else {
					t.graphCreated = true
					t.actions++
				}
			}
			planIndex++

		case 8:
			// drop graphs
			if t.docCollectionCreated && t.numberOfDocsThisDb >= t.MaxDocuments {
				if err := t.dropGraph(t.graphName, false); err != nil {
					t.log.Errorf("Failed to drop graph: %v", err)
				} else {
					t.graphCreated = false
					t.actions++
				}
			}
			planIndex++

		case 9:
			// drop collections
			if t.docCollectionCreated && t.numberOfDocsThisDb >= t.MaxDocuments && !t.graphCreated {
				if err := t.dropCollection(t.docCollectionName); err != nil {
					t.log.Errorf("Failed to drop collection: %v", err)
				} else {
					t.docCollectionCreated = false
					t.numberOfDocsThisDb = 0
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
