package simple

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

func updateExistingDocumentOk(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

	// Expect another try to PATCH:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

	// Expect another try to PATCH:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// this time, let a connection refused happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 1},
		Err:  nil,
	}
	// No GET in this case!

	// Expect another try to PATCH:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// finally, it works:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201, Rev: "abc1236"},
		Err:  nil,
	}

	// Expect another try to PATCH:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// Respond with an unexpected status code, this will be a failure:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  fmt.Errorf("Received unexpected status code 404."),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateExistingDocumentOkWithRetry(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, updateExistingDocumentOk)
	test.client = mockClient
	test.listener = util.MockListener{}
	doc := UserDocument{
		Key:   "doc1",
		Value: 12,
		Name:  "hanswurst",
		Odd:   true,
	}
	// First create a document to update:
	rev, err := test.createDocument(coll, doc, "doc1")
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from createDocument: %v, err: %v", rev, err)
	}
	rev2, err := test.updateExistingDocument(coll, "doc1", rev)
	if rev2 == "" || err != nil || rev2 == rev {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev2, err)
	}
	rev, err = test.updateExistingDocument(coll, "doc1", rev2)
	if rev == "" || err != nil || rev == rev2 {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	rev2, err = test.updateExistingDocument(coll, "doc1", rev)
	if rev2 != "" || err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev2, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentTimeoutOkBehaviour(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}
	// Save new document:
	patch := req.Input.(map[string]interface{})
	newName := patch["name"].(string)

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
		**x = coll.existingDocs["doc1"]
		(*x).Name = newName
		(*x).Rev = "abc1235"
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abc1235"},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateExistingDocumentTimeoutThenOK(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateExistingDocumentTimeoutOkBehaviour)
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
	rev, err = test.updateExistingDocument(coll, "doc1", rev)
	if rev == "" || err != nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentOverallTimeout(
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
		// Get a normal PATCH request:
		select { // here, we do not know if we expect another one or not
		case req = <-requests:
		case <-ctx.Done():
			return
		}
		if req.Method != "PATCH" {
			t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

func TestUpdateExistingDocumentOverallTimeout(t *testing.T) {
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
	mockClient := util.NewMockClient(t, updateExistingDocumentOverallTimeout)
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
	rev, err = test.updateExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentReadTimeout(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

func TestUpdateExistingDocumentReadTimeout(t *testing.T) {
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
	mockClient := util.NewMockClient(t, updateExistingDocumentReadTimeout)
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
	rev, err = test.updateExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentReadNotFound(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

func TestUpdateExistingDocumentReadNotFound(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateExistingDocumentReadNotFound)
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
	rev, err = test.updateExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentReadUnexpected(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

func TestUpdateExistingDocumentReadUnexpected(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateExistingDocumentReadUnexpected)
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
	rev, err = test.updateExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentPreconditionFailed(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateExistingDocumentPreconditionFailed(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateExistingDocumentPreconditionFailed)
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
	rev, err = test.updateExistingDocument(coll, "doc1", rev)
	if rev != "" || err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentPreconditionFailed2(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

	// Get another PATCH request, now round 2:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
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

func TestUpdateExistingDocumentPreconditionFailed2(t *testing.T) {
	stillSeen := false
	butSeen := false
	for {
		test := simpleTest{
			SimpleConfig: config,
			reportDir:    ".",
			log:          log,
			collections:  make(map[string]*collection),
		}
		mockClient := util.NewMockClient(t, updateExistingDocumentPreconditionFailed2)
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
		rev, err = test.updateExistingDocument(coll, "doc1", rev)
		if rev != "" || err == nil {
			t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
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

func updateExistingDocumentWrongRevision(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// return with 412 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1235"},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateExistingDocumentWrongRevision(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateExistingDocumentWrongRevision)
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
	err = test.updateExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla")
	if err != nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	err = test.updateExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2")
	if err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateExistingDocumentWrongRevisionOverallTimeout(
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
		// Get a normal PATCH request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "PATCH" {
			t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestUpdateExistingDocumentWrongRevisionOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateExistingDocumentWrongRevisionOverallTimeout)
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
	err = test.updateExistingDocumentWrongRevision(coll.name, "doc1", rev+"bla2")
	if err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateNonExistingDocument(
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

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// return with 404 precondition failed
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// Get a normal PATCH request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}

	// return with 200
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: "abcd1235"},
		Err:  fmt.Errorf("Test failure"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateNonExistingDocument(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateNonExistingDocument)
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
	err = test.updateNonExistingDocument(coll.name, "doc1")
	if err != nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	err = test.updateNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}

func updateNonExistingDocumentOverallTimeout(
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
		// Get a normal PATCH request:
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "PATCH" {
			t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
		}

		// return with 503
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestUpdateNonExistingDocumentOverallTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, updateNonExistingDocumentOverallTimeout)
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
	err = test.updateNonExistingDocument(coll.name, "doc1")
	if err == nil {
		t.Errorf("Unexpected result from updateExistingDocument: %v, err: %v", rev, err)
	}
	mockClient.Shutdown()
}
