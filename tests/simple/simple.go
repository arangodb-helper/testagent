package simple

import (
	"bytes"
	"fmt"
	"io"
	stdlog "log"
	"math/rand"
	"net/url"
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
	logPath                  string
	reportDir                string
	log                      *logging.Logger
	cluster                  cluster.Cluster
	listener                 test.TestListener
	stop                     chan struct{}
	active                   bool
	client                   *util.ArangoClient
	failures                 int
	actions                  int
	existingDocs             map[string]UserDocument
	readExistingCounter      counter
	readNonExistingCounter   counter
	createCounter            counter
	updateExistingCounter    counter
	updateNonExistingCounter counter
	deleteExistingCounter    counter
	deleteNonExistingCounter counter
	importCounter            counter
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
			fmt.Sprintf("#existing documents removed: %d", t.deleteExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents read: %d", t.readNonExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents updated: %d", t.updateNonExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents removed: %d", t.deleteNonExistingCounter.succeeded),
			fmt.Sprintf("#import operations: %d", t.importCounter.succeeded),
			"",
			"Failed:",
			fmt.Sprintf("#documents created: %d", t.createCounter.failed),
			fmt.Sprintf("#existing documents read: %d", t.readExistingCounter.failed),
			fmt.Sprintf("#existing documents updated: %d", t.updateExistingCounter.failed),
			fmt.Sprintf("#existing documents removed: %d", t.deleteExistingCounter.failed),
			fmt.Sprintf("#non-existing documents read: %d", t.readNonExistingCounter.failed),
			fmt.Sprintf("#non-existing documents updated: %d", t.updateNonExistingCounter.failed),
			fmt.Sprintf("#non-existing documents removed: %d", t.deleteNonExistingCounter.failed),
			fmt.Sprintf("#import operations: %d", t.importCounter.failed),
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
	if err := t.createCollection(collUser, 3, 2); err != nil {
		t.log.Errorf("Failed to create collection (%v). Giving up", err)
		return
	}
	t.actions++

	// Import documents
	if err := t.importDocuments(collUser); err != nil {
		t.log.Errorf("Failed to import documents: %#v", err)
	}
	t.actions++

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
		if err := t.createDocument(collUser, userDoc, userDoc.Key); err != nil {
			t.log.Errorf("Failed to create document: %#v", err)
		}
		t.existingDocs[userDoc.Key] = userDoc
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

	selectRandomKey := func() string {
		index := rand.Intn(len(t.existingDocs))
		for k := range t.existingDocs {
			if index == 0 {
				return k
			}
			index--
		}
		return "" // This should never be reached when len(t.existingDocs) > 0
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
			if err := t.createDocument(collUser, userDoc, userDoc.Key); err != nil {
				t.log.Errorf("Failed to create document: %#v", err)
			} else {
				t.existingDocs[userDoc.Key] = userDoc
			}
			state++

		case 1:
			// Read a random existing document
			if len(t.existingDocs) > 0 {
				randomKey := selectRandomKey()
				if err := t.readExistingDocument(collUser, randomKey); err != nil {
					t.log.Errorf("Failed to read existing document '%s': %#v", randomKey, err)
				}
			}
			state++

		case 2:
			// Read a random non-existing document
			randomKey := createNewKey(false)
			if err := t.readNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to read non-existing document '%s': %#v", randomKey, err)
			}
			state++

		case 3:
			// Remove a random existing document
			if len(t.existingDocs) > 0 {
				randomKey := selectRandomKey()
				if err := t.removeExistingDocument(collUser, randomKey); err != nil {
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

		case 4:
			// Remove a random non-existing document
			randomKey := createNewKey(false)
			if err := t.removeNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to remove non-existing document '%s': %#v", randomKey, err)
			}
			state++

		case 5:
			// Update a random existing document
			if len(t.existingDocs) > 0 {
				randomKey := selectRandomKey()
				if err := t.updateExistingDocument(collUser, randomKey); err != nil {
					t.log.Errorf("Failed to update existing document '%s': %#v", randomKey, err)
				} else {
					// Updated succeeded, now try to read it, it should exist and be updated
					if err := t.readExistingDocument(collUser, randomKey); err != nil {
						t.log.Errorf("Failed to read just-updated document '%s': %#v", randomKey, err)
					}
				}
			}
			state++

		case 6:
			// Update a random non-existing document
			randomKey := createNewKey(false)
			if err := t.updateNonExistingDocument(collUser, randomKey); err != nil {
				t.log.Errorf("Failed to update non-existing document '%s': %#v", randomKey, err)
			}
			state++

		default:
			state = 0
		}

		time.Sleep(time.Second * 5)
	}
}

func (t *simpleTest) createCollection(name string, numberOfShards, replicationFactor int) error {
	opts := struct {
		Name              string `json:"name"`
		NumberOfShards    int    `json:"numberOfShards"`
		ReplicationFactor int    `json:"replicationFactor"`
	}{
		Name:              name,
		NumberOfShards:    numberOfShards,
		ReplicationFactor: replicationFactor,
	}
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d...", name, numberOfShards, replicationFactor)
	if err := t.client.Post("/_api/collection", nil, opts, "", nil, []int{200}, []int{400, 404, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.reportFailure(test.NewFailure("Failed to create collection '%s': %v", name, err))
		return maskAny(err)
	}
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d succeeded", name, numberOfShards, replicationFactor)
	return nil
}

func (t *simpleTest) createDocument(collectionName string, document interface{}, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	t.log.Infof("Creating document '%s' in '%s'...", key, collectionName)
	if err := t.client.Post(fmt.Sprintf("/_api/document/%s", collectionName), q, document, "", nil, []int{200, 201, 202}, []int{400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.createCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create document in collection '%s': %v", collectionName, err))
		return maskAny(err)
	}
	t.createCounter.succeeded++
	t.log.Infof("Creating document '%s' in '%s' succeeded", key, collectionName)
	return nil
}

func (t *simpleTest) readExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	var result UserDocument
	t.log.Infof("Reading existing document '%s' from '%s'...", key, collectionName)
	if err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), nil, &result, []int{200, 201, 202}, []int{400, 404, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.readExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	// Compare document against expected document
	expected := t.existingDocs[key]
	if result.Value != expected.Value || result.Name != expected.Name || result.Odd != expected.Odd {
		// This is a failure
		t.readExistingCounter.failed++
		t.reportFailure(test.NewFailure("Read existing document '%s' returned different values '%s': got %q expected %q", key, collectionName, result, expected))
		return maskAny(fmt.Errorf("Read returned invalid values"))
	}
	t.readExistingCounter.succeeded++
	t.log.Infof("Reading existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}

func (t *simpleTest) readNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	var result UserDocument
	t.log.Infof("Reading non-existing document '%s' from '%s'...", key, collectionName)
	if err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), nil, &result, []int{404}, []int{200, 201, 202, 400, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.readNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.readNonExistingCounter.succeeded++
	t.log.Infof("Reading non-existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}

func (t *simpleTest) updateExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	t.log.Infof("Updating existing document '%s' in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	doc := t.existingDocs[key]
	if err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, delta, "", nil, []int{200, 201, 202}, []int{400, 404, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.updateExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	// Update internal doc
	doc.Name = newName
	t.existingDocs[key] = doc
	t.updateExistingCounter.succeeded++
	t.log.Infof("Updating existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}

func (t *simpleTest) updateNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated non-existing name %s", time.Now())
	t.log.Infof("Updating non-existing document '%s' in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	if err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, delta, "", nil, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.updateNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.updateNonExistingCounter.succeeded++
	t.log.Infof("Updating non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}

func (t *simpleTest) removeExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	t.log.Infof("Removing existing document '%s' from '%s'...", key, collectionName)
	if err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, []int{200, 201, 202}, []int{400, 404, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.deleteExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteExistingCounter.succeeded++
	t.log.Infof("Removing existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}

func (t *simpleTest) removeNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	t.log.Infof("Removing non-existing document '%s' from '%s'...", key, collectionName)
	if err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.deleteNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteNonExistingCounter.succeeded++
	t.log.Infof("Removing non-existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}

func (t *simpleTest) createImportDocument() ([]byte, []UserDocument) {
	buf := &bytes.Buffer{}
	docs := make([]UserDocument, 0, 10000)
	fmt.Fprintf(buf, `[ "_key", "value", "name", "odd" ]`)
	fmt.Fprintln(buf)
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("docimp%05d", i)
		userDoc := UserDocument{
			Key:   key,
			Value: i,
			Name:  fmt.Sprintf("Imported %d", i),
			Odd:   i%2 == 0,
		}
		docs = append(docs, userDoc)
		fmt.Fprintf(buf, `[ "%s", %d, "%s", %v ]`, userDoc.Key, userDoc.Value, userDoc.Name, userDoc.Odd)
		fmt.Fprintln(buf)
	}
	return buf.Bytes(), docs
}

func (t *simpleTest) importDocuments(collectionName string) error {
	operationTimeout, retryTimeout := time.Minute, time.Minute*3
	q := url.Values{}
	q.Set("collection", collectionName)
	q.Set("waitForSync", "true")
	importData, docs := t.createImportDocument()
	t.log.Infof("Importing %d documents ('%s' - '%s') into '%s'...", len(docs), docs[0].Key, docs[len(docs)-1].Key, collectionName)
	if err := t.client.Post("/_api/import", q, importData, "application/x-www-form-urlencoded", nil, []int{200, 201, 202}, []int{400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.importCounter.failed++
		t.reportFailure(test.NewFailure("Failed to import documents in collection '%s': %v", collectionName, err))
		return maskAny(err)
	}
	for _, d := range docs {
		t.existingDocs[d.Key] = d
	}
	t.importCounter.succeeded++
	t.log.Infof("Importing %d documents ('%s' - '%s') into '%s' succeeded", len(docs), docs[0].Key, docs[len(docs)-1].Key, collectionName)
	return nil
}
