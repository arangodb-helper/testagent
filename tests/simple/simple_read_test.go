package simple

import (
	"context"
	"fmt"
	"testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

func readExistingDocumentOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a normal GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_api/document/" + coll.name + "/doc1"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	if x, ok := req.Result.(*UserDocument); ok {
		*x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a second document request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// Respond immediately with a 503:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	// Expect another try to GET:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// this time, let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  nil,
	}

	// Expect another try to GET:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// this time, let a connection refused happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 1},
		Err:  nil,
	}
	// No GET in this case!

	// Expect another try to GET:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// finally, it works:
	if x, ok := req.Result.(*UserDocument); ok {
		*x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abc1234"},
		Err:  nil,
	}

	// Expect another try to GET:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// Respond with an unexpected status code, this will be a failure:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  fmt.Errorf("Received unexpected status code 404."),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDocumentOkWithRetry(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, readExistingDocumentOk)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	// First create a document to read:
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev2, err := test.readExistingDocument(coll, "doc1", rev, false, false)
	if rev2 == "" || err != nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev2, err)
	}
	rev, err = test.readExistingDocument(coll, "doc1", rev2, false, false)
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	rev2, err = test.readExistingDocument(coll, "doc1", rev, false, false)
	if rev2 != "" || err == nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev2, err)
	}
	mockClient.Shutdown()
}

func readExistingDocumentTimeoutOkBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a normal GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  nil,
	}

	// Expect another GET request to see if the document is there, answer
	// with yes:
	path = "/_api/document/" + coll.name + "/doc1"
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Respond with found:
	if x, ok := req.Result.(*UserDocument); ok {
		*x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abc1234"},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDocumentTimeoutThenOK(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readExistingDocumentTimeoutOkBehaviour)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev2, err := test.readExistingDocument(coll, "doc1", rev, false, false)
	if rev2 == "" || err != nil || rev != rev2 {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func readExistingDocumentOverallTimeout(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	for {
		// Get a normal GET request:
		select { // here, we do not know if we expect another one or not
		case req = <-requests:
		case <-ctx.Done():
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}

		// let a timeout happen:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 0},
			Err:  nil,
		}
	}
}

func TestReadExistingDocumentOverallTimeout(t *testing.T) {
	saveReadTimeout := ReadTimeout
	ReadTimeout = 5 // to speed up timeout failure, needs to be longer than
	// operationTimeout*4, which is 4
	defer func() { ReadTimeout = saveReadTimeout }()

	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readExistingDocumentOverallTimeout)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev, err = test.readExistingDocument(coll, "doc1", rev, false, false)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func readExistingDocumentReadUnexpected(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a read requests:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// Return with an unknown document, which is a failure:
	strange := UserDocument{
		Name:  "Strange",
		Value: 4711,
		Odd:   true,
		Key:   "doc1",
		Rev:   "grzfzl",
	}
	if x, ok := req.Result.(*UserDocument); ok {
		*x = strange
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "grzfzl"},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDocumentReadUnexpected(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readExistingDocumentReadUnexpected)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev, err = test.readExistingDocument(coll, "doc1", rev, false, false)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func readExistingDocumentPreconditionFailed(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a normal GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  fmt.Errorf("Unexpected status code 412"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDocumentPreconditionFailed(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readExistingDocumentPreconditionFailed)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev, err = test.readExistingDocument(coll, "doc1", rev, false, false)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func readExistingDocumentWrongRevision(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a normal GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// Get a normal GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1235"},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDocumentWrongRevision(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readExistingDocumentWrongRevision)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	err = test.readExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla", false)
	if err != nil {
		t.Errorf("Unexpected result from readExistingDocumentWrongRevision: %v, err: %v", rev, err)
	}
	err = test.readExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2", false)
	if err == nil {
		t.Errorf("Unexpected result from readExistingDocumentWrongRevision: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func readExistingDocumentWrongRevisionOverallTimeout(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	for {
		// Get a normal GET request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestReadExistingDocumentWrongRevisionOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readExistingDocumentWrongRevisionOverallTimeout)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	err = test.readExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2", false)
	if err == nil {
		t.Errorf("Unexpected result from readExistingDocumentWrongRevision: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func readNonExistingDocument(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a normal GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// return with 404 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// Get a normal GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1235"},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadNonExistingDocument(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readNonExistingDocument)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	err = test.readNonExistingDocument(coll.name, "doc1")
	if err != nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	err = test.readNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func readNonExistingDocumentOverallTimeout(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	for {
		// Get a normal GET request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestReadNonExistingDocumentOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, readNonExistingDocumentOverallTimeout)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	err = test.readNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from readExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}
