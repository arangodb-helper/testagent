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
	NumberOfShards    int
	ReplicationFactor int
	OperationTimeout  time.Duration
	RetryTimeout      time.Duration
}

type Replication2TestHarness struct {
	activeMutex sync.Mutex
	logPath     string
	reportDir   string
	log         *logging.Logger
	cluster     cluster.Cluster
	listener    test.TestListener
	client      util.ArangoClientInterface
}

type Replication2Counters struct {
	createDatabaseCounter     counter
	createCollectionCounter   counter
	createGraphCounter        counter
	dropCollectionCounter     counter
	dropDatabaseCounter       counter
	dropGraphCounter          counter
	singleDocCreateCounter    counter
	edgeDocumentCreateCounter counter
	bulkCreateCounter         counter
	readExistingCounter       counter
	updateExistingCounter     counter
	replaceExistingCounter    counter
	traverseGraphCounter      counter
	queryCreateCursorCounter  counter
}

type Replication2TestContext struct {
	Replication2Config
	Replication2TestHarness
	Replication2Counters
	documentIdSeq     int64
	collectionNameSeq int64
	existingDocuments []TestDocument
}

type Replication2Test struct {
	Replication2TestContext
	TestName       string
	stop           chan struct{}
	active         bool
	pauseRequested bool
	paused         bool
	failures       int
	actions        int
}

type counter struct {
	succeeded int
	failed    int
}

var (
	ReadTimeout int = 128 // to be overwritten in unittests only
)

// Name returns the name of the script
func (t *Replication2Test) Name() string {
	return t.TestName
}

// Stop any running test. This should not return until tests are actually stopped.
func (t *Replication2Test) Stop() error {
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
func (t *Replication2Test) Pause() error {
	t.pauseRequested = true
	return nil
}

// Resume running the tests, where Pause interrupted it.
func (t *Replication2Test) Resume() error {
	t.pauseRequested = false
	return nil
}

// CollectLogs copies all logging info to the given writer.
func (t *Replication2Test) CollectLogs(w io.Writer) error {
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
func (t *Replication2Test) setupLogger(cluster cluster.Cluster) error {
	t.logPath = filepath.Join(t.reportDir, fmt.Sprintf("%s-%s.log", t.Name(), cluster.ID()))
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

func (t *Replication2Test) shouldStop() bool {
	// Should we stop?
	if stop := t.stop; stop != nil {
		stop <- struct{}{}
		return true
	}
	return false
}

func (t *Replication2Test) reportFailure(f test.Failure) {
	t.failures++
	t.listener.ReportFailure(f)
}
