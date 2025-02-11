package simple

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

func replaceExistingDocumentOk(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}
	path = "/_api/document/" + coll.name + "/doc1"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1235"},
		Err:  nil,
	}

	// Get a second document request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abc1235"},
		Err:  nil,
	}

	// Expect another try to PUT:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// this time, answer with 410:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}

	// Expect another GET request to see if the document is there, answer
	// with old document:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Respond with found, but original version:
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abc1234"},
		Err:  nil,
	}

	// Expect another try to PUT:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Respond with found, but original version:
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abc1234"},
		Err:  nil,
	}

	// Expect another try to PUT:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// this time, let a connection refused happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 1},
		Err:  nil,
	}
	// No GET in this case!

	// Expect another try to PUT:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// finally, it works:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201, Rev: "abc1236"},
		Err:  nil,
	}

	// Expect another try to PUT:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// Respond with an unexpected status code, this will be a failure:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  fmt.Errorf("Received unexpected status code 404."),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceExistingDocumentOkWithRetry(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, replaceExistingDocumentOk)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	// First create a document to replace:
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev2, err := test.replaceExistingDocument(coll, "doc1", rev)
	if rev2 == "" || err != nil || rev2 == rev {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev2, err)
	}
	rev, err = test.replaceExistingDocument(coll, "doc1", rev2)
	if rev == "" || err != nil || rev == rev2 {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	rev2, err = test.replaceExistingDocument(coll, "doc1", rev)
	if rev2 != "" || err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev2, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentTimeoutOkBehaviour(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}
	// Save new document:
	newDoc := req.Input.(UserDocument)

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
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = newDoc
		(*x).Rev = "abc1235"
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abc1235"},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceExistingDocumentTimeoutThenOK(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceExistingDocumentTimeoutOkBehaviour)
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
	rev, err = test.replaceExistingDocument(coll, "doc1", rev)
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentOverallTimeout(
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
		// Get a normal PUT request:
		select { // here, we do not know if we expect another one or not
		case req = <-requests:
		case <-ctx.Done():
			return
		}
		if req.Method != "PUT" {
			t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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

func TestReplaceExistingDocumentOverallTimeout(t *testing.T) {
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
	mockClient := util.NewMockClient(t, replaceExistingDocumentOverallTimeout)
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
	rev, err = test.replaceExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentReadTimeout(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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

func TestReplaceExistingDocumentReadTimeout(t *testing.T) {
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
	mockClient := util.NewMockClient(t, replaceExistingDocumentReadTimeout)
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
	rev, err = test.replaceExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentReadNotFound(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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

func TestReplaceExistingDocumentReadNotFound(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceExistingDocumentReadNotFound)
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
	rev, err = test.replaceExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentGoneButWrittenFailBehaviour(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}
	newDocument := req.Input.(UserDocument)

	// respond with 410:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
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

	if x, ok := req.Result.(**UserDocument); ok {
		*x = &newDocument
	}

	// Return new document:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceExistingDocumentGoneButWritten(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceExistingDocumentGoneButWrittenFailBehaviour)
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
	_, err = test.replaceExistingDocument(coll, "doc1", rev)
	if err == nil {
		t.Error("Unexpected result from replaceExistingDocument. Exprected an error.")
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentReadUnexpected(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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

	// Return with an unknown document, which is a failure:
	strange := UserDocument{
		Name:  "Strange",
		Value: 4711,
		Odd:   true,
		Key:   "doc1",
		Rev:   "grzfzl",
	}
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &strange
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceExistingDocumentReadUnexpected(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceExistingDocumentReadUnexpected)
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
	rev, err = test.replaceExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentPreconditionFailed(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceExistingDocumentPreconditionFailed(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceExistingDocumentPreconditionFailed)
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
	rev, err = test.replaceExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentPreconditionFailed2(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// first a timeout:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  nil,
	}

	// Expect a GET request to see if the document is there, answer
	// with yes and give the old one:
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

	// Get another PUT request, now round 2:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// return with 412 precondition failed:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// This could now either run into a failure if no if-match header
	// was given, or it could retry, in which case another GET comes
	// our way:

	// Expect a GET request to see if the document is there, answer
	// with yes and give the old one:
	req = potentialNext(ctx, t, requests)
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

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceExistingDocumentPreconditionFailed2(t *testing.T) {
	stillSeen := false
	butSeen := false
	for {
		test := simpleTest{
			SimpleConfig: config,
			reportDir:    ".",
			log:          log,
			collections:  make(map[string]*collection),
		}
		mockClient := util.NewMockClient(t, replaceExistingDocumentPreconditionFailed2)
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
		rev, err = test.replaceExistingDocument(coll, "doc1", rev)
		if rev != "" || err == nil {
			t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
		}
		if strings.Contains(err.Error(), "but did not set") {
			butSeen = true
		}
		if strings.Contains(err.Error(), "still") {
			stillSeen = true
		}
		if butSeen && stillSeen {
			break
		}

		mockClient.Shutdown()
	}
}

func replaceExistingDocumentWrongRevision(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1235"},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceExistingDocumentWrongRevision(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceExistingDocumentWrongRevision)
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
	err = test.replaceExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla")
	if err != nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	err = test.replaceExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2")
	if err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceExistingDocumentWrongRevisionOverallTimeout(
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
		// Get a normal PUT request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "PUT" {
			t.Errorf("Got wrong method %s instead of PUT.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestReplaceExistingDocumentWrongRevisionOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceExistingDocumentWrongRevisionOverallTimeout)
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
	err = test.replaceExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2")
	if err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceNonExistingDocument(
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

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// return with 404 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// Get a normal PUT request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1235"},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReplaceNonExistingDocument(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceNonExistingDocument)
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
	err = test.replaceNonExistingDocument(coll.name, "doc1")
	if err != nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	err = test.replaceNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func replaceNonExistingDocumentOverallTimeout(
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
		// Get a normal PUT request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "PUT" {
			t.Errorf("Got wrong method %s instead of PUT.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestReplaceNonExistingDocumentOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, replaceNonExistingDocumentOverallTimeout)
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
	err = test.replaceNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from replaceExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}
