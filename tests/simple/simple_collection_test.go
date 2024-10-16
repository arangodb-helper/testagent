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
		existingDocs: map[string]UserDocument{},
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
	mockClient.Shutdown()
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
	mockClient.Shutdown()
}

func createCollectionTestTimeoutFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request (as preparation)
	var req *util.MockRequest

	for {
		req = potentialNext(ctx, t, requests)
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
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
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
	mockClient.Shutdown()
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
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}
		// Respond with temporarily unavailable:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse { StatusCode: 503, },
			Err:  nil,
		}
	}

}

func TestCollectionCreateReadTimeoutFail(t *testing.T) {
	ReadTimeout = 5 // to speed up timeout failure, needs to be longer than
	// operationTimeout*4, which is 4
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
	mockClient.Shutdown()
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
	// Answer with temporarily not available:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 503, },
		Err:  nil,
	}

	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	// Answer with a error:
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
	mockClient.Shutdown()
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
	mockClient.Shutdown()
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
	// Expect a try to GET
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
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
	mockClient.Shutdown()
}

func createCollectionGoneOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	//POST -> 410 -> GET -> 404 -> POST -> 200
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// answer with 410:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}
	// Expect a try to GET
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
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
	// answer with 200:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionCreateGoneOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionGoneOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err != nil {
		t.Errorf("Unexpected result from createCollection: err: %v", err)
	}
	mockClient.Shutdown()
}

func createCollectionGoneButCreatedFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	//POST -> 410 -> GET -> 200 -> fail
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	// answer with 410:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}
	// Expect a try to GET
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateCollectionGoneButCreatedFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, createCollectionGoneButCreatedFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.createCollection(foo, 9, 2)
	if err == nil {
		t.Errorf("Unexpected result from createCollection: expected an error!")
	}
	mockClient.Shutdown()
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
	// Expect a try to POST again:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
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
	mockClient.Shutdown()
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
	mockClient.Shutdown()
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
	mockClient.Shutdown()
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
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
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
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
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
	mockClient.Shutdown()
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
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
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
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
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
	mockClient.Shutdown()
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
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
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
	mockClient.Shutdown()
}

func removeExistingCollectionErrorFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
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
	mockClient.Shutdown()
}

func removeExistingCollectionMissingFirstFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
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
	mockClient.Shutdown()
}

func removeExistingCollectionGoneThenOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}
	path := "/_api/collection/" + foo.name
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCollectionExistingRemoveGoneThenOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, removeExistingCollectionGoneThenOk)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Remove a collection
	err := test.removeExistingCollection(foo)
	if err != nil {
		t.Errorf("Unexpected result from removeExistingCollection: err: %v", err)
	}
	mockClient.Shutdown()
}

func removeExistingCollectionMissingLaterOk(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
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
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
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
	mockClient.Shutdown()
}

func removeExistingCollectionTimeoutFail(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	var req *util.MockRequest

	for {
		req = potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "DELETE" {
			t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
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
	ReadTimeout = 5 // to speed up timeout failure, needs to be longer than
	// operationTimeout*4, which is 4
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
	mockClient.Shutdown()
}

