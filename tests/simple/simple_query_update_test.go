package simple

import (
	"context"
	"encoding/json"
	"fmt"
//	"math/rand"
	"testing"
//	"time"

	"github.com/arangodb-helper/testagent/tests/util"
)

func queryUpdateOK(
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
	path := "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 201 },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUpdateOK(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateOK)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocuments(foo, "key")
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateUnavailable(
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
	path := "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 503 },
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
	path = "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 201 },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUpdateUnavailable(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateUnavailable)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocuments(foo, "key")
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateFail(
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
	path := "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 404 },
		Err:  fmt.Errorf("Fail"),
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUpdateFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocuments(foo, "key")
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateResultCountOff(
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
	path := "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	json.Unmarshal([]byte(`{
  "result": [
    {
      "_key": "10044",
      "_id": "foo/10044",
      "_rev": "_cUDCmV6---",
      "i": 100000
    },
    {
      "_key": "10045",
      "_id": "foo/10045",
      "_rev": "_cUDCm6V---",
      "i": 100000
    }
  ],
  "hasMore": false,
  "cached": false,
  "extra": {
    "warnings": [],
    "stats": {
      "writesExecuted": 1,
      "writesIgnored": 0,
      "scannedFull": 0,
      "scannedIndex": 1,
      "filtered": 0,
      "httpRequests": 0,
      "executionTime": 0.003047824007808231,
      "peakMemoryUsage": 0
    }
  },
  "error": false,
  "code": 201
  }`), &req.Result)
	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 201 },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUpdateResultCountOff(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateResultCountOff)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocuments(foo, "key")
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateTimeout(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	for {
		// Get a normal POST request (as preparation)
		req := potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "POST" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
		}
		path := "/_api/cursor"
		if req.UrlPath != path {
			t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
		}

		// Answer with a normal good response: created.
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{ StatusCode: 503 },
			Err:  nil,
		}
	}
	
}

func TestQueryUpdateTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateTimeout)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocuments(foo, "key")
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateLongRunningOK(
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
	path := "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 201 },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUpdateLongRunningOK(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateLongRunningOK)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocumentsLongRunning(foo, "key")
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateLongRunningFail(
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
	path := "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 404 },
		Err:  fmt.Errorf("Fail"),
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUpdateLongRunningFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateLongRunningFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocumentsLongRunning(foo, "key")
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateLongRunningResultCountOff(
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
	path := "/_api/cursor"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}

	json.Unmarshal([]byte(`{
  "result": [
    {
      "_key": "10044",
      "_id": "foo/10044",
      "_rev": "_cUDCmV6---",
      "i": 100000
    },
    {
      "_key": "10045",
      "_id": "foo/10045",
      "_rev": "_cUDCm6V---",
      "i": 100000
    }
  ],
  "hasMore": false,
  "cached": false,
  "extra": {
    "warnings": [],
    "stats": {
      "writesExecuted": 1,
      "writesIgnored": 0,
      "scannedFull": 0,
      "scannedIndex": 1,
      "filtered": 0,
      "httpRequests": 0,
      "executionTime": 0.003047824007808231,
      "peakMemoryUsage": 0
    }
  },
  "error": false,
  "code": 201
  }`), &req.Result)
	// Answer with a normal good response: created.
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{ StatusCode: 201 },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUpdateLongRunningResultCountOff(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateLongRunningResultCountOff)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocumentsLongRunning(foo, "key")
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUpdateLongRunningTimeout(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	for {
		// Get a normal POST request (as preparation)
		req := potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "POST" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
		}
		path := "/_api/cursor"
		if req.UrlPath != path {
			t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
		}

		// Answer with a normal good response: created.
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{ StatusCode: 503 },
			Err:  nil,
		}
	}
	
}

func TestQueryUpdateLongRunningTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUpdateLongRunningTimeout)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	_, err := test.queryUpdateDocumentsLongRunning(foo, "key")
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

