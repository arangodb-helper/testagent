package simple

import (
	"context"
	"fmt"
	"testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

func removeExistingDocumentOk(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}
	path = "/_api/document/" + coll.name + "/doc1"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// Get a normal POST request (as preparation)
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a second document request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// Respond immediately with a 503:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	// Now expect a GET request to see if the document is there, answer
	// with yes:
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

	// Respond with found, but original version:
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// Get another DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// Respond immediately with a 410:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}

	// Now expect a GET request to see if the document is there, answer
	// with yes:
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

	// Respond with found, but original version:
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// Expect another try to DELETE:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// this time, let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  nil,
	}

	// Expect another GET request to see if the document is there, answer
	// with old document:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}

	// Respond with found, but original version:
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// Expect another try to DELETE:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// this time, let a connection refused happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 1},
		Err:  nil,
	}
	// No GET in this case!

	// Expect another try to DELETE:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// finally, it works:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// Get a normal POST request (as preparation)
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/document/" + coll.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Expect another try to DELETE:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// Respond with an unexpected status code, this will be a failure:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  fmt.Errorf("Received unexpected status code 404."),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestRemoveExistingDocumentOkWithRetry(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, removeExistingDocumentOk)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	// First create a document to remove:
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err != nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	rev, err = test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err != nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	rev, err = test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentTimeoutOkBehaviour(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
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

	// Respond with not found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestRemoveExistingDocumentTimeoutThenOK(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeExistingDocumentTimeoutOkBehaviour)
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
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err != nil {
		t.Errorf("Unexpected result from removeExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentOverallTimeout(
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
		// Get a normal DELETE request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "DELETE" {
			t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
		}

		// let a timeout happen:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 0},
			Err:  nil,
		}

		// Expect another GET request to see if the document is there, answer
		// with no:
		req = next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}

		// Respond with old document:
		if x, ok := req.Result.(**UserDocument); ok {
			*x = &UserDocument{}
			**x = coll.existingDocs["doc1"]
		}
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 200},
			Err:  nil,
		}
	}
}

func TestRemoveExistingDocumentOverallTimeout(t *testing.T) {
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
	mockClient := util.NewMockClient(t, removeExistingDocumentOverallTimeout)
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
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentReadTimeout(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  nil,
	}

	for {
		// Get a sequence of read requests:
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

func TestRemoveExistingDocumentReadTimeout(t *testing.T) {
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
	mockClient := util.NewMockClient(t, removeExistingDocumentReadTimeout)
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
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentReadNotFound(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
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

	// Return with not found, which is a failure:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestRemoveExistingDocumentReadNotFound(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeExistingDocumentReadNotFound)
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
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err != nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentNotFound(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// respond with not found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
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

	// Respond with found, but original version:
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1234"},
		Err:  nil,
	}

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// respond with 404 not found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestRemoveExistingDocumentNotFound(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeExistingDocumentNotFound)
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
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err != nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentPreconditionFailed(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  fmt.Errorf("unexpected code 412"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestRemoveExistingDocumentPreconditionFailed(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeExistingDocumentPreconditionFailed)
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
	err = test.removeExistingDocument(coll.name, "doc1", rev)
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument, err: %v", err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentWrongRevision(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestRemoveExistingDocumentWrongRevision(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeExistingDocumentWrongRevision)
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
	err = test.removeExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla")
	if err != nil {
		t.Errorf("Unexpected result from removeExistingDocument: %v, err: %v", rev, err)
	}
	err = test.removeExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2")
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func removeExistingDocumentWrongRevisionOverallTimeout(
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
		// Get a normal DELETE request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "DELETE" {
			t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestRemoveExistingDocumentWrongRevisionOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeExistingDocumentWrongRevisionOverallTimeout)
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
	err = test.removeExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2")
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func removeNonExistingDocument(
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

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// return with 404 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// Get a normal DELETE request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestRemoveNonExistingDocument(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeNonExistingDocument)
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
	err = test.removeNonExistingDocument(coll.name, "doc1")
	if err != nil {
		t.Errorf("Unexpected result from removeExistingDocument: %v, err: %v", rev, err)
	}
	err = test.removeNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func removeNonExistingDocumentOverallTimeout(
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
		// Get a normal DELETE request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "DELETE" {
			t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestRemoveNonExistingDocumentOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, removeNonExistingDocumentOverallTimeout)
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
	err = test.removeNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from removeExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}
