package simple

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/url"
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
	log                      *logging.Logger
	cluster                  cluster.Cluster
	listener                 test.TestListener
	stop                     chan struct{}
	active                   bool
	client                   *util.ArangoClient
	failures                 int
	existingDocs             map[string]UserDocument
	readExistingCounter      counter
	readNonExistingCounter   counter
	createCounter            counter
	deleteExistingCounter    counter
	deleteNonExistingCounter counter
	importCounter            counter
}

type counter struct {
	succeeded int
	failed    int
}

// NewSimpleTest creates a simple test
func NewSimpleTest(log *logging.Logger) test.TestScript {
	return &simpleTest{
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
		Messages: []string{
			fmt.Sprintf("Current #documents: %d", len(t.existingDocs)),
			"Succeeded:",
			fmt.Sprintf("#documents created: %d", t.createCounter.succeeded),
			fmt.Sprintf("#existing documents read: %d", t.readExistingCounter.succeeded),
			fmt.Sprintf("#existing documents removed: %d", t.deleteExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents read: %d", t.readNonExistingCounter.succeeded),
			fmt.Sprintf("#non-existing documents removed: %d", t.deleteNonExistingCounter.succeeded),
			fmt.Sprintf("#import operations: %d", t.importCounter.succeeded),
			"",
			"Failed:",
			fmt.Sprintf("#documents created: %d", t.createCounter.failed),
			fmt.Sprintf("#existing documents read: %d", t.readExistingCounter.failed),
			fmt.Sprintf("#existing documents removed: %d", t.deleteExistingCounter.failed),
			fmt.Sprintf("#non-existing documents read: %d", t.readNonExistingCounter.failed),
			fmt.Sprintf("#non-existing documents removed: %d", t.deleteNonExistingCounter.failed),
			fmt.Sprintf("#import operations: %d", t.importCounter.failed),
		},
	}
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
	if err := t.createCollection(collUser, 3, 2); err != nil {
		t.log.Errorf("Failed to create collection (%v). Giving up", err)
		return
	}

	// Import documents
	t.log.Debugf("Importing documents")
	if err := t.importDocuments(collUser); err != nil {
		t.log.Errorf("Failed to import documents: %#v", err)
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
		t.log.Debugf("Trying to create document %#v", userDoc)
		if err := t.createDocument(collUser, userDoc); err != nil {
			t.log.Errorf("Failed to create document: %#v", err)
		}
		t.existingDocs[userDoc.Key] = userDoc
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

		switch state {
		case 0:
			// Create a random document
			userDoc := UserDocument{
				Key:   createNewKey(true),
				Value: rand.Int(),
				Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
				Odd:   time.Now().Nanosecond()%2 == 1,
			}
			if err := t.createDocument(collUser, userDoc); err != nil {
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
	timeout := time.Minute
	if err := t.client.Post("/_api/collection", nil, opts, "", nil, []int{200}, []int{400, 404, 307}, timeout); err != nil {
		// This is a failure
		t.reportFailure(test.NewFailure("Failed to create collection '%s': %v", name, err))
		return maskAny(err)
	}
	return nil
}

func (t *simpleTest) createDocument(collectionName string, document interface{}) error {
	timeout := time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	if err := t.client.Post(fmt.Sprintf("/_api/document/%s", collectionName), q, document, "", nil, []int{200, 201, 202}, []int{400, 404, 409, 307}, timeout); err != nil {
		// This is a failure
		t.createCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create document in collection '%s': %v", collectionName, err))
		return maskAny(err)
	}
	t.createCounter.succeeded++
	return nil
}

func (t *simpleTest) readExistingDocument(collectionName string, key string) error {
	timeout := time.Minute
	var result UserDocument
	if err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), nil, &result, []int{200, 201, 202}, []int{400, 404, 307}, timeout); err != nil {
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
	return nil
}

func (t *simpleTest) readNonExistingDocument(collectionName string, key string) error {
	timeout := time.Minute
	var result UserDocument
	if err := t.client.Get(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), nil, &result, []int{404}, []int{200, 201, 202, 400, 307}, timeout); err != nil {
		// This is a failure
		t.readNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.readNonExistingCounter.succeeded++
	return nil
}

func (t *simpleTest) removeExistingDocument(collectionName string, key string) error {
	timeout := time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	if err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, []int{200, 201, 202}, []int{400, 404, 412, 307}, timeout); err != nil {
		// This is a failure
		t.deleteExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteExistingCounter.succeeded++
	return nil
}

func (t *simpleTest) removeNonExistingDocument(collectionName string, key string) error {
	timeout := time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	if err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, []int{404}, []int{200, 201, 202, 400, 412, 307}, timeout); err != nil {
		// This is a failure
		t.deleteNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteNonExistingCounter.succeeded++
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
	timeout := time.Minute
	q := url.Values{}
	q.Set("collection", collectionName)
	q.Set("waitForSync", "true")
	importData, docs := t.createImportDocument()
	if err := t.client.Post("/_api/import", q, importData, "application/x-www-form-urlencoded", nil, []int{200, 201, 202}, []int{400, 404, 409, 307}, timeout); err != nil {
		// This is a failure
		t.importCounter.failed++
		t.reportFailure(test.NewFailure("Failed to import documents in collection '%s': %v", collectionName, err))
		return maskAny(err)
	}
	for _, d := range docs {
		t.existingDocs[d.Key] = d
	}
	t.importCounter.succeeded++
	return nil
}
