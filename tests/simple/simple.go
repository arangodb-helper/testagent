package simple

import (
	"fmt"
	"io"
	stdlog "log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/arangodb/testAgent/service/cluster"
	"github.com/arangodb/testAgent/service/test"
	"github.com/arangodb/testAgent/tests/util"
	logging "github.com/op/go-logging"
)

const (
	collUser             = "simple_users"
	initialDocumentCount = 999
)

type simpleTest struct {
	logPath                             string
	reportDir                           string
	log                                 *logging.Logger
	cluster                             cluster.Cluster
	listener                            test.TestListener
	stop                                chan struct{}
	active                              bool
	client                              *util.ArangoClient
	failures                            int
	actions                             int
	existingDocs                        map[string]UserDocument
	readExistingCounter                 counter
	readExistingWrongRevisionCounter    counter
	readNonExistingCounter              counter
	createCounter                       counter
	updateExistingCounter               counter
	updateExistingWrongRevisionCounter  counter
	updateNonExistingCounter            counter
	replaceExistingCounter              counter
	replaceExistingWrongRevisionCounter counter
	replaceNonExistingCounter           counter
	deleteExistingCounter               counter
	deleteExistingWrongRevisionCounter  counter
	deleteNonExistingCounter            counter
	importCounter                       counter
	queryCreateCursorCounter            counter
	queryNextBatchCounter               counter
	queryNextBatchNewCoordinatorCounter counter
	queryLongRunningCounter             counter
}

type counter struct {
	succeeded int
	failed    int
}

// NewSimpleTest creates a simple test
func NewSimpleTest(log *logging.Logger, reportDir string) test.TestScript {
	return &simpleTest{
		reportDir:    reportDir,
		log:          log,
		existingDocs: make(map[string]UserDocument),
	}
}

// Name returns the name of the script
func (t *simpleTest) Name() string {
	return "simple"
}

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *simpleTest) Start(cluster cluster.Cluster, listener test.TestListener) error {
	if err := t.setupLogger(cluster); err != nil {
		return maskAny(err)
	}

	t.cluster = cluster
	t.listener = listener
	t.client = util.NewArangoClient(t.log, cluster)

	go t.testLoop()
	return nil
}

// Stop any running test. This should not return until tests are actually stopped.
func (t *simpleTest) Stop() error {
	stop := make(chan struct{})
	t.stop = stop
	<-stop
	return nil
}

// Status returns the current status of the test
func (t *simpleTest) Status() test.TestStatus {
	return test.TestStatus{
		Failures: t.failures,
		Actions:  t.actions,
		Messages: []string{
			fmt.Sprintf("Current #documents: %d", len(t.existingDocs)),
			"Succeeded:",
			fmt.Sprintf("#documents created: %d", t.createCounter.succeeded),
			fmt.Sprintf("#existing documents read: %d", t.readExistingCounter.succeeded),
			fmt.Sprintf("#existing documents updated: %d", t.updateExistingCounter.succeeded),
			fmt.Sprintf("#existing documents replaced: %d", t.replaceExistingCounter.succeeded),
			fmt.Sprintf("#existing documents removed: %d", t.deleteExistingCounter.succeeded),
			fmt.Sprintf("#existing documents wrong revision read: %d", t.readExistingWrongRevisionCounter.succeeded),
			fmt.Sprintf("#existing documents wrong revision updated: %d", t.updateExistingWrongRevisionCounter.succeeded),
			fmt.Sprintf("#existing documents wrong revision replaced: %d", t.replaceExistingWrongRevisionCounter.succeeded),
			fmt.Sprintf("#existing documents wrong revision removed: %d", t.deleteExistingWrongRevisionCounter.succeeded),
			fmt.Sprintf("#non-existing documents read: %d", t.readNonExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents updated: %d", t.updateNonExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents replaced: %d", t.replaceNonExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents removed: %d", t.deleteNonExistingCounter.succeeded),
			fmt.Sprintf("#import operations: %d", t.importCounter.succeeded),
			fmt.Sprintf("#create AQL cursor operations: %d", t.queryCreateCursorCounter.succeeded),
			fmt.Sprintf("#fetch next AQL cursor batch operations: %d", t.queryNextBatchCounter.succeeded),
			fmt.Sprintf("#fetch next AQL cursor batch after coordinator change operations: %d", t.queryNextBatchNewCoordinatorCounter.succeeded),
			fmt.Sprintf("#long running AQL query operations: %d", t.queryLongRunningCounter.succeeded),
			"",
			"Failed:",
			fmt.Sprintf("#documents created: %d", t.createCounter.failed),
			fmt.Sprintf("#existing documents read: %d", t.readExistingCounter.failed),
			fmt.Sprintf("#existing documents updated: %d", t.updateExistingCounter.failed),
			fmt.Sprintf("#existing documents replaced: %d", t.replaceExistingCounter.failed),
			fmt.Sprintf("#existing documents removed: %d", t.deleteExistingCounter.failed),
			fmt.Sprintf("#existing documents wrong revision read: %d", t.readExistingWrongRevisionCounter.failed),
			fmt.Sprintf("#existing documents wrong revision updated: %d", t.updateExistingWrongRevisionCounter.failed),
			fmt.Sprintf("#existing documents wrong revision replaced: %d", t.replaceExistingWrongRevisionCounter.failed),
			fmt.Sprintf("#existing documents wrong revision removed: %d", t.deleteExistingWrongRevisionCounter.failed),
			fmt.Sprintf("#non-existing documents read: %d", t.readNonExistingCounter.failed),
			fmt.Sprintf("#non-existing documents updated: %d", t.updateNonExistingCounter.failed),
			fmt.Sprintf("#non-existing documents replaced: %d", t.replaceNonExistingCounter.failed),
			fmt.Sprintf("#non-existing documents removed: %d", t.deleteNonExistingCounter.failed),
			fmt.Sprintf("#import operations: %d", t.importCounter.failed),
			fmt.Sprintf("#create AQL cursor operations: %d", t.queryCreateCursorCounter.failed),
			fmt.Sprintf("#fetch next AQL cursor batch operations: %d", t.queryNextBatchCounter.failed),
			fmt.Sprintf("#fetch next AQL cursor batch after coordinator change operations: %d", t.queryNextBatchNewCoordinatorCounter.failed),
			fmt.Sprintf("#long running AQL query operations: %d", t.queryLongRunningCounter.failed),
		},
	}
}

// CollectLogs copies all logging info to the given writer.
func (t *simpleTest) CollectLogs(w io.Writer) error {
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
func (t *simpleTest) setupLogger(cluster cluster.Cluster) error {
	t.logPath = filepath.Join(t.reportDir, fmt.Sprintf("simple-%s.log", cluster.ID()))
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

func (t *simpleTest) shouldStop() bool {
	// Should we stop?
	if stop := t.stop; stop != nil {
		stop <- struct{}{}
		return true
	}
	return false
}

type UserDocument struct {
	Key   string `json:"_key"`
	rev   string // Note that we do not export this field!
	Value int    `json:"value"`
	Name  string `json:"name"`
	Odd   bool   `json:"odd"`
}

func (t *simpleTest) reportFailure(f test.Failure) {
	t.failures++
	t.listener.ReportFailure(f)
}

func (t *simpleTest) testLoop() {
	t.active = true
	defer func() { t.active = false }()

	t.existingDocs = make(map[string]UserDocument)
	t.actions = 0
	if err := t.createCollection(collUser, 9, 2); err != nil {
		t.log.Errorf("Failed to create collection (%v). Giving up", err)
		return
	}
	t.actions++

	// Import documents
	if err := t.importDocuments(collUser); err != nil {
		t.log.Errorf("Failed to import documents: %#v", err)
	}
	t.actions++

	// Check imported documents
	for k := range t.existingDocs {
		if t.shouldStop() {
			return
		}
		if err := t.readExistingDocument(collUser, k, "", true); err != nil {
			t.log.Errorf("Failed to read existing document '%s': %#v", k, err)
		}
		t.actions++
	}

	// Create sample users
	for i := 0; i < initialDocumentCount; i++ {
		if t.shouldStop() {
			return
		}
		userDoc := UserDocument{
			Key:   fmt.Sprintf("doc%05d", i),
			Value: i,
			Name:  fmt.Sprintf("User %d", i),
			Odd:   i%2 == 1,
		}
		if rev, err := t.createDocument(collUser, userDoc, userDoc.Key); err != nil {
			t.log.Errorf("Failed to create document: %#v", err)
		} else {
			userDoc.rev = rev
			t.existingDocs[userDoc.Key] = userDoc
		}
		t.actions++
	}

	createNewKey := func(record bool) string {
		for {
			key := fmt.Sprintf("newkey%07d", rand.Int31n(100*1000))
			_, found := t.existingDocs[key]
			if !found {
				if record {
					t.existingDocs[key] = UserDocument{}
				}
				return key
			}
		}
	}

	removeExistingKey := func(key string) {
		delete(t.existingDocs, key)
	}

	selectRandomKey := func() (string, string) {
		index := rand.Intn(len(t.existingDocs))
		for k, v := range t.existingDocs {
			if index == 0 {
				return k, v.rev
			}
			index--
		}
		return "", "" // This should never be reached when len(t.existingDocs) > 0
	}

	selectWrongRevision := func(key string) (string, bool) {
		correctRev := t.existingDocs[key].rev
		for _, v := range t.existingDocs {
			if v.rev != correctRev && v.rev != "" {
				return v.rev, true
			}
		}
		return "", false // This should never be reached when len(t.existingDocs) > 1
	}

	state := 0
	for {
		// Should we stop
		if t.shouldStop() {
			return
		}
		t.actions++

		switch state {
		case 0:
			// Create a random document
			userDoc := UserDocument{
				Key:   createNewKey(true),
				Value: rand.Int(),
				Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
				Odd:   time.Now().Nanosecond()%2 == 1,
			}
			if rev, err := t.createDocument(collUser, userDoc, userDoc.Key); err != nil {
				t.log.Errorf("Failed to create document: %#v", err)
			} else {
				userDoc.rev = rev
				t.existingDocs[userDoc.Key] = userDoc

				// Now try to read it, it must exist
				if err := t.readExistingDocument(collUser, userDoc.Key, rev, false); err != nil {
					t.log.Errorf("Failed to read just-created document '%s': %#v", userDoc.Key, err)
				}
			}
			state++

		case 1:
			// Read a random existing document
			if len(t.existingDocs) > 0 {
				randomKey, rev := selectRandomKey()
				if err := t.readExistingDocument(collUser, randomKey, rev, false); err != nil {
					t.log.Errorf("Failed to read existing document '%s': %#v", randomKey, err)
				}
			}
			state++

		case 2:
			// Read a random existing document but with wrong revision
			if len(t.existingDocs) > 1 {
				randomKey, _ := selectRandomKey()
				if rev, ok := selectWrongRevision(randomKey); ok {
					if err := t.readExistingDocumentWrongRevision(collUser, randomKey, rev, false); err != nil {
						t.log.Errorf("Failed to read existing document '%s' wrong revision: %#v", randomKey, err)
					}
				}
			}
			state++

		case 3:
			// Read a random non-existing document
			randomKey := createNewKey(false)
			if err := t.readNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to read non-existing document '%s': %#v", randomKey, err)
			}
			state++

		case 4:
			// Remove a random existing document
			if len(t.existingDocs) > 0 {
				randomKey, rev := selectRandomKey()
				if err := t.removeExistingDocument(collUser, randomKey, rev); err != nil {
					t.log.Errorf("Failed to remove existing document '%s': %#v", randomKey, err)
				} else {
					// Remove succeeded, key should no longer exist
					removeExistingKey(randomKey)

					// Now try to read it, it should not exist
					if err := t.readNonExistingDocument(collUser, randomKey); err != nil {
						t.log.Errorf("Failed to read just-removed document '%s': %#v", randomKey, err)
					}
				}
			}
			state++

		case 5:
			// Remove a random existing document but with wrong revision
			if len(t.existingDocs) > 1 {
				randomKey, correctRev := selectRandomKey()
				if rev, ok := selectWrongRevision(randomKey); ok {
					if err := t.removeExistingDocumentWrongRevision(collUser, randomKey, rev); err != nil {
						t.log.Errorf("Failed to remove existing document '%s' wrong revision: %#v", randomKey, err)
					} else {
						// Remove failed (as expected), key should still exist
						if err := t.readExistingDocument(collUser, randomKey, correctRev, false); err != nil {
							t.log.Errorf("Failed to read not-just-removed document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			state++

		case 6:
			// Remove a random non-existing document
			randomKey := createNewKey(false)
			if err := t.removeNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to remove non-existing document '%s': %#v", randomKey, err)
			}
			state++

		case 7:
			// Update a random existing document
			if len(t.existingDocs) > 0 {
				randomKey, rev := selectRandomKey()
				if newRev, err := t.updateExistingDocument(collUser, randomKey, rev); err != nil {
					t.log.Errorf("Failed to update existing document '%s': %#v", randomKey, err)
				} else {
					// Updated succeeded, now try to read it, it should exist and be updated
					if err := t.readExistingDocument(collUser, randomKey, newRev, false); err != nil {
						t.log.Errorf("Failed to read just-updated document '%s': %#v", randomKey, err)
					}
				}
			}
			state++

		case 8:
			// Update a random existing document but with wrong revision
			if len(t.existingDocs) > 1 {
				randomKey, correctRev := selectRandomKey()
				if rev, ok := selectWrongRevision(randomKey); ok {
					if err := t.updateExistingDocumentWrongRevision(collUser, randomKey, rev); err != nil {
						t.log.Errorf("Failed to update existing document '%s' wrong revision: %#v", randomKey, err)
					} else {
						// Updated failed (as expected).
						// It must still be readable.
						if err := t.readExistingDocument(collUser, randomKey, correctRev, false); err != nil {
							t.log.Errorf("Failed to read not-just-updated document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			state++

		case 9:
			// Update a random non-existing document
			randomKey := createNewKey(false)
			if err := t.updateNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to update non-existing document '%s': %#v", randomKey, err)
			}
			state++

		case 10:
			// Replace a random existing document
			if len(t.existingDocs) > 0 {
				randomKey, rev := selectRandomKey()
				if newRev, err := t.replaceExistingDocument(collUser, randomKey, rev); err != nil {
					t.log.Errorf("Failed to replace existing document '%s': %#v", randomKey, err)
				} else {
					// Replace succeeded, now try to read it, it should exist and be replaced
					if err := t.readExistingDocument(collUser, randomKey, newRev, false); err != nil {
						t.log.Errorf("Failed to read just-replaced document '%s': %#v", randomKey, err)
					}
				}
			}
			state++

		case 11:
			// Replace a random existing document but with wrong revision
			if len(t.existingDocs) > 1 {
				randomKey, correctRev := selectRandomKey()
				if rev, ok := selectWrongRevision(randomKey); ok {
					if err := t.replaceExistingDocumentWrongRevision(collUser, randomKey, rev); err != nil {
						t.log.Errorf("Failed to replace existing document '%s' wrong revision: %#v", randomKey, err)
					} else {
						// Replace failed (as expected).
						// It must still be readable.
						if err := t.readExistingDocument(collUser, randomKey, correctRev, false); err != nil {
							t.log.Errorf("Failed to read not-just-replaced document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			state++

		case 12:
			// Replace a random non-existing document
			randomKey := createNewKey(false)
			if err := t.replaceNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to replace non-existing document '%s': %#v", randomKey, err)
			}
			state++

		case 13:
			// Query documents
			if err := t.queryDocuments(collUser); err != nil {
				t.log.Errorf("Failed to query documents: %#v", err)
			}
			state++

		case 14:
			// Query documents (long running)
			if err := t.queryDocumentsLongRunning(collUser); err != nil {
				t.log.Errorf("Failed to query (long running) documents: %#v", err)
			}
			state++

		default:
			state = 0
		}

		time.Sleep(time.Second * 5)
	}
}

// createRandomIfMatchHeader creates a request header with one of the following (randomly chosen):
// 1: with an `If-Match` entry for the given revision.
// 2: without an `If-Match` entry for the given revision.
func createRandomIfMatchHeader(hdr map[string]string, rev string) (map[string]string, string) {
	if rev == "" {
		return hdr, "without If-Match"
	}
	switch rand.Intn(2) {
	case 0:
		hdr = ifMatchHeader(hdr, rev)
		return hdr, "with If-Match"
	default:
		return hdr, "without If-Match"
	}
}

// ifMatchHeader creates a request header with an `If-Match` entry for the given revision.
func ifMatchHeader(hdr map[string]string, rev string) map[string]string {
	if rev == "" {
		panic(fmt.Errorf("rev cannot be empty"))
	}
	if hdr == nil {
		hdr = make(map[string]string)
	}
	hdr["If-Match"] = rev
	return hdr
}
