package replication2

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

type Replication2Config struct {
	MaxDocuments      int
	BatchSize         int
	DocumentSize      int
	NumberOfShards    int
	ReplicationFactor int
	OperationTimeout  time.Duration
	RetryTimeout      time.Duration
}

type replication2Test struct {
	Replication2Config
	activeMutex              sync.Mutex
	logPath                  string
	reportDir                string
	log                      *logging.Logger
	cluster                  cluster.Cluster
	listener                 test.TestListener
	stop                     chan struct{}
	active                   bool
	pauseRequested           bool
	paused                   bool
	lastRequestErr           bool
	client                   util.ArangoClientInterface
	failures                 int
	actions                  int
	collectionName           string
	collectionCreated        bool
	numberOfDocsThisDb       int
	numberOfCreatedDocsTotal int64
	documentIdSeq            int64
	existingDocSeeds         []int64
	createCollectionCounter  counter
	dropCollectionCounter    counter
	bulkCreateCounter        counter
	readExistingCounter      counter
}

type counter struct {
	succeeded int
	failed    int
}

// NewReplication2Test creates a replication2 test
func NewReplication2Test(log *logging.Logger, reportDir string, config Replication2Config) test.TestScript {
	return &replication2Test{
		Replication2Config: config,
		reportDir:          reportDir,
		log:                log,
		collectionCreated:  false,
		documentIdSeq:      0,
		existingDocSeeds:   make([]int64, config.MaxDocuments),
		collectionName:     "replication2_docs",
	}
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
			plan = []int{0, 1, 2, 3, 4, 5} // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
			// create a collection
			if !t.collectionCreated {
				if err := t.createCollection(t.collectionName); err != nil {
					t.log.Errorf("Failed to create collection: %v", err)
				} else {
					t.collectionCreated = true
					t.actions++
				}
			}
			planIndex++

		case 1:
			// create documents
			if t.collectionCreated {
				for {
					if t.numberOfDocsThisDb >= t.MaxDocuments {
						break
					}
					var thisBatchSize int
					if t.BatchSize <= t.MaxDocuments-t.numberOfDocsThisDb {
						thisBatchSize = t.BatchSize
					} else {
						thisBatchSize = t.MaxDocuments - t.numberOfDocsThisDb
					}
					if err := t.createDocuments(thisBatchSize, t.documentIdSeq); err != nil {
						t.log.Errorf("Failed to create documents: %#v", err)
					} else {
						t.numberOfDocsThisDb += thisBatchSize
						t.numberOfCreatedDocsTotal += int64(thisBatchSize)
					}
					t.documentIdSeq += int64(thisBatchSize)
					t.actions++
				}
			}
			planIndex++

		case 2:
			// read documents
			if t.collectionCreated {
				for _, seed := range t.existingDocSeeds {
					if err := t.readExistingDocument(t.collectionName, seed, false); err != nil {
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
			planIndex++

		case 4:
			// read documents again after update
			planIndex++

		case 5:
			// drop collection
			if t.collectionCreated && t.numberOfDocsThisDb >= t.MaxDocuments {
				if err := t.dropCollection(t.collectionName); err != nil {
					t.log.Errorf("Failed to drop collection: %v", err)
				} else {
					t.collectionCreated = false
					t.numberOfDocsThisDb = 0
					t.existingDocSeeds = make([]int64, t.MaxDocuments)
					t.dropCollectionCounter.succeeded++
					t.actions++
				}
			}
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}
