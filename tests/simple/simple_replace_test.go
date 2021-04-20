package simple

import (
	"context"
	"fmt"
  "testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

func replaceExistingDocumentOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abcd1234" },
		Err: nil,
	}

	// Get a normal PUT request:
	req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}
	path = "/_api/document/" + coll.name + "/doc1"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abcd1235" },
		Err: nil,
	}

	// Get a second document request:
	req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// Respond immediately with a 503:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 503 },
		Err: nil,
	}

	// Now expect a GET request to see if the document is there, answer
	// with yes:
	req = next(ctx, t, requests, true); if req == nil { return }
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
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abc1235" },
		Err:  nil,
	}

	// Expect another try to PUT:
	req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// this time, let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 0 },
		Err:  nil,
	}

	// Expect another GET request to see if the document is there, answer
	// with old document:
	req = next(ctx, t, requests, true); if req == nil { return }
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Respond with found, but original version:
	if x, ok := req.Result.(**UserDocument); ok {
		*x = &UserDocument{}
		**x = coll.existingDocs["doc1"]
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abc1234" },
		Err:  nil,
	}

	// Expect another try to PUT:
	req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// this time, let a connection refused happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 1 },
		Err:  nil,
	}
	// No GET in this case!

	// Expect another try to PUT:
	req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// finally, it works:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 201, Rev: "abc1236" },
		Err:  nil,
	}

	// Expect another try to PUT:
	req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}

	// Respond with an unexpected status code, this will be a failure:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 404 },
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
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
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
  req := next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abcd1234" },
		Err: nil,
	}

	// Get a normal PUT request:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
	}
	// Save new document:
	newDoc := req.Input.(UserDocument)

	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 0 },
		Err:  nil,
	}

	// Expect another GET request to see if the document is there, answer
	// with yes:
	path = "/_api/document/" + coll.name + "/doc1";
  req = next(ctx, t, requests, true); if req == nil { return }
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
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abc1235" },
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
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
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
  req := next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abcd1234" },
		Err: nil,
	}

	for {
		// Get a normal PUT request:
		select {  // here, we do not know if we expect another one or not
		case req = <-requests:
		case <-ctx.Done():
			return
		}
		if req.Method != "PUT" {
			t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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
		if x, ok := req.Result.(**UserDocument); ok {
			*x = &UserDocument{}
		  **x = coll.existingDocs["doc1"]
		}
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{ StatusCode: 200},
			Err: nil,
		}
	}
}

func TestReplaceExistingDocumentOverallTimeout(t *testing.T) {
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
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
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
  req := next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/document/" + coll.name;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 200, Rev: "abcd1234" },
		Err: nil,
	}

	// Get a normal PUT request:
  req = next(ctx, t, requests, true); if req == nil { return }
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of PUT.", req.Method)
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

func TestReplaceExistingDocumentReadTimeout(t *testing.T) {
	ReadTimeout = 5   // to speed up timeout failure, needs to be longer than
	                  // operationTimeout*4, which is 4
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
    Key: "doc1",
		Value: 12,
		Name: "hanswurst",
		Odd: true,
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

