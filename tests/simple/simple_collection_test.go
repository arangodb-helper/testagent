package simple

import (
	"context"
	"fmt"
	"testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

var (
	foo *collection = &collection {
		name: "foo",
	}
)

func createCollectionOk(
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
	path := "/_api/collection"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response:
	// Respond with found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err != nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func createCollectionRedirectFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{},
		Err: fmt.Errorf("Error"),
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateRedirectFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionRedirectFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func createCollectionTestTimeoutFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	var req *util.MockRequest

	for {
		req = next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method != "POST" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
		}
		// Answer with a normal good response:
		// Respond with found:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse { StatusCode: 503, },
			Err:  nil,
		}
		// Get a normal POST request (as preparation)
		req = next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
		}
		// Respond with not found:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse { StatusCode: 404, },
			Err:  nil,
		}
	}

}

func TestCollectionCreateTestTimeoutFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionTestTimeoutFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func createCollectionReadTimeoutFail(
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
	// Answer with a normal good response:
	// Respond with found:
	responses <- &util.MockResponse{
			Resp: util.ArangoResponse { StatusCode: 503, },
		Err:  nil,
	}
	for {
		// Get a normal POST request (as preparation)
		req = next(ctx, t, requests, true)
		if req == nil {
			return
		}
		// Answer with a normal good response:
		// Respond with found:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse { StatusCode: 503, },
			Err:  nil,
		}
	}

}

func TestCollectionCreateReadTimeoutFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionReadTimeoutFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func createCollectionCreateReadErrorFail(
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
	// Answer with a normal good response:
	// Respond with found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 503, },
		Err:  nil,
	}

	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// Answer with a normal good response:
	// Respond with found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {},
		Err:  fmt.Errorf("error"),
	}

	// Get a normal POST request (as preparation)
	req = next(ctx, t, requests, true)
	if req == nil {
		return 
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// Answer with a normal good response:
	// Respond with found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 409, },
		Err:  nil,
	}

	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// Answer with a normal good response:
	// Respond with found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {},
		Err:  fmt.Errorf("error"),
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateReadErrorFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionCreateReadErrorFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func createCollectionTimeoutOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// POST -> timeout -> retry -> 200
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  nil,
	}
	// Expect another try to POST
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}


func TestCollectionCreateTimeoutOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionTimeoutOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err != nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}

}

func createCollectionUnknownOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	//POST -> 503 -> retry -> 200
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}
	// Expect another try to POST
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateUnknownOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionUnknownOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err != nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}

}


func createCollectionRefusedExistsFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// POST -> 1 -> 200 -> error
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 1},
		Err:  nil,
	}
	// Expect another try to POST
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateRefusedExistsFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionRefusedExistsFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}

}

func createCollectionUnfinishedExistsFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// POST -> 1 -> 200 -> error
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 500},
		Err:  nil,
	}
	// Expect another try to POST
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateUnfinishedExistsFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionUnfinishedExistsFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}

}


func createCollectionConflictFirstAttemptFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateConflictFirstAttemptFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionConflictFirstAttemptFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}

}


func createCollectionConflictLaterAttemptExistsOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// respond with not found:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// respond with conflict:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// return OK:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateConflictLaterAttemptExistsOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionConflictLaterAttemptExistsOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection -> 409 -> 200 must fail
	err := test.createCollection(foo, 9, 2)
	if err != nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}

}

func createCollectionConflictLaterAttemptNotExistsFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// let a timeout happen:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateConflictLaterAttemptNotExistsFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionConflictLaterAttemptNotExistsFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection -> 409 -> 200 must fail
	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}

}


func removeExistingCollectionOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/collection/" + foo.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionExistingRemoveOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, removeExistingCollectionOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.removeExistingCollection(foo)
	if err != nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func removeExistingCollectionErrorFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/collection/" + foo.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{},
		Err:  fmt.Errorf("error"),
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionExistingRemoveErrorFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, removeExistingCollectionErrorFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.removeExistingCollection(foo)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func removeExistingCollectionMissingFirstFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/collection/" + foo.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 404, },
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionExistingRemoveMissingFirstFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, removeExistingCollectionMissingFirstFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.removeExistingCollection(foo)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func removeExistingCollectionMissingLaterOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path := "/_api/collection/" + foo.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 503, },
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 404, },
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionExistingRemoveMissingLaterOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, removeExistingCollectionMissingLaterOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.removeExistingCollection(foo)
	if err != nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

func removeExistingCollectionTimeoutFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	var req *util.MockRequest

	for {
		req = next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method != "DELETE" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
		}
		path := "/_api/collection/" + foo.name
		if req.UrlPath != path {
			t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
		}
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse { StatusCode: 503, },
			Err:  nil,
		}
	}
}

func TestCollectionExistingRemoveTimeoutFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, removeExistingCollectionTimeoutFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.removeExistingCollection(foo)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
}

