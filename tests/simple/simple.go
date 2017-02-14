package simple

import (
	"fmt"
	"io"
	stdlog "log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/arangodb/testAgent/service/cluster"
	"github.com/arangodb/testAgent/service/test"
	"github.com/arangodb/testAgent/tests/util"
	logging "github.com/op/go-logging"
)

type SimpleConfig struct {
	MaxDocuments int
}

const (
	collUser             = "simple_users"
	initialDocumentCount = 999
)

type simpleTest struct {
	SimpleConfig
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
	rebalanceShardsCounter              counter
	queryUpdateCounter                  counter
}

type counter struct {
	succeeded int
	failed    int
}

// NewSimpleTest creates a simple test
func NewSimpleTest(log *logging.Logger, reportDir string, config SimpleConfig) test.TestScript {
	return &simpleTest{
		SimpleConfig: config,
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
	cc := func(name string, c counter) test.Counter {
		return test.Counter{
			Name:      name,
			Succeeded: c.succeeded,
			Failed:    c.failed,
		}
	}

	return test.TestStatus{
		Failures: t.failures,
		Actions:  t.actions,
		Messages: []string{
			fmt.Sprintf("Current #documents: %d", len(t.existingDocs)),
		},
		Counters: []test.Counter{
			cc("#documents created", t.createCounter),
			cc("#existing documents read", t.readExistingCounter),
			cc("#existing documents updated", t.updateExistingCounter),
			cc("#existing documents replaced", t.replaceExistingCounter),
			cc("#existing documents removed", t.deleteExistingCounter),
			cc("#existing documents wrong revision read", t.readExistingWrongRevisionCounter),
			cc("#existing documents wrong revision updated", t.updateExistingWrongRevisionCounter),
			cc("#existing documents wrong revision replaced", t.replaceExistingWrongRevisionCounter),
			cc("#existing documents wrong revision removed", t.deleteExistingWrongRevisionCounter),
			cc("#non-existing documents read", t.readNonExistingCounter),
			cc("#non-existing documents updated", t.updateNonExistingCounter),
			cc("#non-existing documents replaced", t.replaceNonExistingCounter),
			cc("#non-existing documents removed", t.deleteNonExistingCounter),
			cc("#import operations", t.importCounter),
			cc("#create AQL cursor operations", t.queryCreateCursorCounter),
			cc("#fetch next AQL cursor batch operations", t.queryNextBatchCounter),
			cc("#fetch next AQL cursor batch after coordinator change operations", t.queryNextBatchNewCoordinatorCounter),
			cc("#long running AQL query operations", t.queryLongRunningCounter),
			cc("#rebalance shards operations", t.rebalanceShardsCounter),
			cc("#update AQL query operations", t.queryUpdateCounter),
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

	var plan []int
	planIndex := 0
	for {
		// Should we stop
		if t.shouldStop() {
			return
		}
		t.actions++
		if plan == nil || planIndex >= len(plan) {
			plan = createTestPlan(17) // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
			// Create a random document
			if len(t.existingDocs) < t.MaxDocuments {
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
					t.client.SetCoordinator("")
					if err := t.readExistingDocument(collUser, userDoc.Key, rev, false); err != nil {
						t.log.Errorf("Failed to read just-created document '%s': %#v", userDoc.Key, err)
					}
				}
			}
			planIndex++

		case 1:
			// Read a random existing document
			if len(t.existingDocs) > 0 {
				randomKey, rev := selectRandomKey()
				if err := t.readExistingDocument(collUser, randomKey, rev, false); err != nil {
					t.log.Errorf("Failed to read existing document '%s': %#v", randomKey, err)
				}
			}
			planIndex++

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
			planIndex++

		case 3:
			// Read a random non-existing document
			randomKey := createNewKey(false)
			if err := t.readNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to read non-existing document '%s': %#v", randomKey, err)
			}
			planIndex++

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
					t.client.SetCoordinator("")
					if err := t.readNonExistingDocument(collUser, randomKey); err != nil {
						t.log.Errorf("Failed to read just-removed document '%s': %#v", randomKey, err)
					}
				}
			}
			planIndex++

		case 5:
			// Remove a random existing document but with wrong revision
			if len(t.existingDocs) > 1 {
				randomKey, correctRev := selectRandomKey()
				if rev, ok := selectWrongRevision(randomKey); ok {
					if err := t.removeExistingDocumentWrongRevision(collUser, randomKey, rev); err != nil {
						t.log.Errorf("Failed to remove existing document '%s' wrong revision: %#v", randomKey, err)
					} else {
						// Remove failed (as expected), key should still exist
						t.client.SetCoordinator("")
						if err := t.readExistingDocument(collUser, randomKey, correctRev, false); err != nil {
							t.log.Errorf("Failed to read not-just-removed document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++

		case 6:
			// Remove a random non-existing document
			randomKey := createNewKey(false)
			if err := t.removeNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to remove non-existing document '%s': %#v", randomKey, err)
			}
			planIndex++

		case 7:
			// Update a random existing document
			if len(t.existingDocs) > 0 {
				randomKey, rev := selectRandomKey()
				if newRev, err := t.updateExistingDocument(collUser, randomKey, rev); err != nil {
					t.log.Errorf("Failed to update existing document '%s': %#v", randomKey, err)
				} else {
					// Updated succeeded, now try to read it, it should exist and be updated
					t.client.SetCoordinator("")
					if err := t.readExistingDocument(collUser, randomKey, newRev, false); err != nil {
						t.log.Errorf("Failed to read just-updated document '%s': %#v", randomKey, err)
					}
				}
			}
			planIndex++

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
						t.client.SetCoordinator("")
						if err := t.readExistingDocument(collUser, randomKey, correctRev, false); err != nil {
							t.log.Errorf("Failed to read not-just-updated document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++

		case 9:
			// Update a random non-existing document
			randomKey := createNewKey(false)
			if err := t.updateNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to update non-existing document '%s': %#v", randomKey, err)
			}
			planIndex++

		case 10:
			// Replace a random existing document
			if len(t.existingDocs) > 0 {
				randomKey, rev := selectRandomKey()
				if newRev, err := t.replaceExistingDocument(collUser, randomKey, rev); err != nil {
					t.log.Errorf("Failed to replace existing document '%s': %#v", randomKey, err)
				} else {
					// Replace succeeded, now try to read it, it should exist and be replaced
					t.client.SetCoordinator("")
					if err := t.readExistingDocument(collUser, randomKey, newRev, false); err != nil {
						t.log.Errorf("Failed to read just-replaced document '%s': %#v", randomKey, err)
					}
				}
			}
			planIndex++

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
						t.client.SetCoordinator("")
						if err := t.readExistingDocument(collUser, randomKey, correctRev, false); err != nil {
							t.log.Errorf("Failed to read not-just-replaced document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++

		case 12:
			// Replace a random non-existing document
			randomKey := createNewKey(false)
			if err := t.replaceNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to replace non-existing document '%s': %#v", randomKey, err)
			}
			planIndex++

		case 13:
			// Query documents
			if err := t.queryDocuments(collUser); err != nil {
				t.log.Errorf("Failed to query documents: %#v", err)
			}
			planIndex++

		case 14:
			// Query documents (long running)
			if err := t.queryDocumentsLongRunning(collUser); err != nil {
				t.log.Errorf("Failed to query (long running) documents: %#v", err)
			}
			planIndex++

		case 15:
			// Rebalance shards
			if err := t.rebalanceShards(); err != nil {
				t.log.Errorf("Failed to rebalance shards: %#v", err)
			}
			planIndex++

		case 16:
			// AQL update query
			if len(t.existingDocs) > 0 {
				randomKey, _ := selectRandomKey()
				if newRev, err := t.queryUpdateDocuments(collUser, randomKey); err != nil {
					t.log.Errorf("Failed to update document using AQL query: %#v", err)
				} else {
					// Updated succeeded, now try to read it (anywhere), it should exist and be updated
					t.client.SetCoordinator("")
					if err := t.readExistingDocument(collUser, randomKey, newRev, false); err != nil {
						t.log.Errorf("Failed to read just-updated document '%s': %#v", randomKey, err)
					}
				}
			}
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}

// createTestPlan creates an int-array of 'steps' long with all values from 0..steps-1 in random order.
func createTestPlan(steps int) []int {
	plan := make([]int, steps)
	for i := 0; i < steps; i++ {
		plan[i] = i
	}
	util.Shuffle(sort.IntSlice(plan))
	return plan
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
