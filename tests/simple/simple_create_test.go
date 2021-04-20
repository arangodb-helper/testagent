package simple

import (
	"context"
	"fmt"
	"encoding/json"
  "testing"
  "time"
	logging "github.com/op/go-logging"

	"github.com/arangodb-helper/testagent/service"
	"github.com/arangodb-helper/testagent/service/cluster/arangodb"
	"github.com/arangodb-helper/testagent/tests/util"
)

var (
  log = logging.MustGetLogger("testAgentTests")
	appFlags struct {
		port int
    service.ServiceConfig
    arangodb.ArangodbConfig
    SimpleConfig
    logLevel string
  }
  config SimpleConfig = SimpleConfig{
		MaxDocuments:     20000,
		MaxCollections:   10,
		OperationTimeout: time.Second * 1,
		RetryTimeout:     time.Minute * 4,
	}
	coll *collection = &collection{
		name: "simple_test_collection",
    existingDocs: make(map[string]UserDocument, 1000),
	}
)

func next(ctx context.Context, t *testing.T, requests chan *util.MockRequest, expectMore bool) *util.MockRequest {
	select {
	case req := <-requests:
		if !expectMore {
			t.Errorf("Did not expect further request.")
		}
		return req
	case <-ctx.Done():
		if expectMore {
      t.Errorf("Expecting further requests.")
		}
		return nil
	}
}

func createDocumentOkBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request:
  req := next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/document/" + coll.name {
		t.Errorf("Got wrong URL path %s instead of /_api/document/%s",
		         req.UrlPath, coll.name)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abcd1234" },
		Err: nil,
	}

	// Get a second document request:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}

	// Respond immediately with a 503:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 503 },
		Err: nil,
	}

	// Now expect a GET request to see if the document is there, answer
	// with no:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	if req.UrlPath != "/_api/document/" + coll.name + "/doc2" {
		t.Errorf("Got wrong URL path %s instead of /_api/document/%s/doc2", req.UrlPath, coll.name)
	}

	// Respond with not found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 404 },
		Err:  nil,
	}

	// Expect another try to POST:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}

	// this time, let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 0 },
		Err:  nil,
	}

	// Expect another GET request to see if the document is there, answer
	// with no:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	if req.UrlPath != "/_api/document/" + coll.name + "/doc2" {
		t.Errorf("Got wrong URL path %s instead of /_api/document/%s/doc2", req.UrlPath, coll.name)
	}

	// Respond with not found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 404 },
		Err:  nil,
	}

	// Expect another try to POST:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}

	// this time, let a connection refused happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 1 },
		Err:  nil,
	}
	// No GET in this case!

	// Expect another try to POST:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}

	// finally, it works:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 201, Rev: "abc123" },
		Err:  nil,
	}

	// Expect another try to POST: (doc3 now)
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}

	// Respond with an unexpected status code, this will be a failure:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 412 },
		Err:  fmt.Errorf("Received unexpected status code 412."),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateDocumentOkWithRetry(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, createDocumentOkBehaviour)
  test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
  }
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev, err = test.createDocument(coll, doc, "doc2")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev, err = test.createDocument(coll, doc, "doc3")
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
  mockClient.Shutdown()
}

func createDocumentTimeoutOkBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request:
  req := next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}

	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 0 },
		Err:  nil,
	}

	// Expect another GET request to see if the document is there, answer
	// with yes:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	if req.UrlPath != "/_api/document/" + coll.name + "/doc1" {
		t.Errorf("Got wrong URL path %s instead of /_api/document/%s/doc2", req.UrlPath, coll.name)
	}

	// Respond with found:
	json.Unmarshal([]byte(`{"name":"hanswurst", "value": 12, "odd": true, "_key": "doc1", "_rev": "abc12345"}`),
	               req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abc12345" },
		Err:  nil,
	}


	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateDocumentTimeoutThenOK(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, createDocumentTimeoutOkBehaviour)
  test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
  }
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
  mockClient.Shutdown()
}

func createDocumentOverallTimeout(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

  var req *util.MockRequest
	for {
		// Get a normal POST request:
		select {  // here, we do not know if we expect another one or not
		case req = <-requests:
		case <-ctx.Done():
			return
		}
		if req.Method != "POST" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
		}

		// let a timeout happen:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{ StatusCode: 0 },
			Err:  nil,
		}

		// Expect another GET request to see if the document is there, answer
		// with no:
		req = next(ctx, t, requests, true); if req == nil { return }
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}

		// Respond with not found:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{ StatusCode: 404 },
			Err:  nil,
		}
	}
}

func TestCreateDocumentOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, createDocumentOverallTimeout)
  test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
  }
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
  mockClient.Shutdown()
}

func createDocumentReadTimeout(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request:
  req := next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}

	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 0 },
		Err:  nil,
	}

	for {
		// Get a sequence of read requests:
		select {  // here, we do not know if we expect another one or not
		case req = <-requests:
		case <-ctx.Done():
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}

		// let a timeout happen:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{ StatusCode: 0 },
			Err:  nil,
		}
	}
}

func TestCreateDocumentReadTimeout(t *testing.T) {
	ReadTimeout = 5   // to speed up timeout failure, needs to be longer than
	                  // operationTimeout*4, which is 4
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, createDocumentReadTimeout)
  test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
  }
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
  mockClient.Shutdown()
}

