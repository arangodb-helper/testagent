package simple

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/arangodb-helper/testagent/tests/util"
)

func queryDocumentsEmptyCollection(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	next(ctx, t, requests, false)
}

func TestQueryDocumentsEmptyCollection(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, queryDocumentsEmptyCollection)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Create collection
	err := test.queryDocuments(foo)
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryDocumentsCursorFail(
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
		Resp: util.ArangoResponse{},
		Err:  fmt.Errorf("oops"),
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryDocumentsCursorFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	mockClient := util.NewMockClient(t, queryDocumentsCursorFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPostUnavailable(
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

func TestQueryPostUnavailable(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryPostUnavailable)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPostTimeout(
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
			Resp: util.ArangoResponse{ StatusCode: 0 },
			Err:  nil,
		}
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	//next(ctx, t, requests, false)
}

func TestQueryPostTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryPostTimeout)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryDocumentsCursorLengthZero(
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
		Resp: util.ArangoResponse { StatusCode: 201, },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryDocumentsCursorLengthZero(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryDocumentsCursorLengthZero)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryDocumentsCursorHasMore(
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
	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 201, CoordinatorURL: "https://coordinator:8529"},
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 1}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, CoordinatorURL: "https://coordinator:8529"},
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = false
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 1}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, CoordinatorURL: "https://coordinator:8529"},
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryDocumentsCursorMasMore(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryDocumentsCursorHasMore)
	test.client = mockClient
	test.listener = util.MockListener{}

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPutFail(
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
	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 201, CoordinatorURL: "https://coordinator:8529"},
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {},
		Err:  fmt.Errorf("PUT Error"),
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryPutFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryPutFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryUptimeFail(
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
	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 201, CoordinatorURL: "https://coordinator:8529"},
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = false
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 1}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {},
		Err:  fmt.Errorf("Error"),
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryUptimeFail(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryUptimeFail)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryWrongCoordinator(
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
	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 201, CoordinatorURL: "https://coordinator:8529"},
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = false
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 0}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://other-coordinator:8529" },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryWrongCoordinator(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryWrongCoordinator)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPutNotFound(
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
	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = false
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 10}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryPutNotFound(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	mockClient := util.NewMockClient(t, queryPutNotFound)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPutNotFoundShortUptime(
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
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{
  "result": [
    {
      "_key": "13360",
      "_id": "foo/13360",
      "_rev": "_cTyP566---",
      "i": 3
    },
    {
      "_key": "13365",
      "_id": "foo/13365",
      "_rev": "_cTyP57----",
      "i": 8
    },
    {
      "_key": "13367",
      "_id": "foo/13367",
      "_rev": "_cTyP57---A",
      "i": 10
    },
    {
      "_key": "13368",
      "_id": "foo/13368",
      "_rev": "_cTyP57---B",
      "i": 11
    },
    {
      "_key": "13372",
      "_id": "foo/13372",
      "_rev": "_cTyP57---C",
      "i": 15
    },
    {
      "_key": "13373",
      "_id": "foo/13373",
      "_rev": "_cTyP57C---",
      "i": 16
    },
    {
      "_key": "13374",
      "_id": "foo/13374",
      "_rev": "_cTyP57C--_",
      "i": 17
    },
    {
      "_key": "13376",
      "_id": "foo/13376",
      "_rev": "_cTyP57C--A",
      "i": 19
    },
    {
      "_key": "13377",
      "_id": "foo/13377",
      "_rev": "_cTyP57C--B",
      "i": 20
    },
    {
      "_key": "13382",
      "_id": "foo/13382",
      "_rev": "_cTyP57C--C",
      "i": 25
    }
  ],
  "hasMore": true,
  "cached": false,
  "extra": {
    "warnings": [],
    "stats": {
      "writesExecuted": 0,
      "writesIgnored": 0,
      "scannedFull": 20,
      "scannedIndex": 0,
      "filtered": 0,
      "httpRequests": 6,
      "executionTime": 0.003913449996616691,
      "peakMemoryUsage": 0
    }
  },
  "error": false,
  "code": 201
}`), &req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = false
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 404, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 0}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryPutNotFoundShortUptime(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	mockClient := util.NewMockClient(t, queryPutNotFoundShortUptime)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPutNotFoundLongUptime(
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
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{
  "result": [
    {
      "_key": "13360",
      "_id": "foo/13360",
      "_rev": "_cTyP566---",
      "i": 3
    },
    {
      "_key": "13365",
      "_id": "foo/13365",
      "_rev": "_cTyP57----",
      "i": 8
    },
    {
      "_key": "13367",
      "_id": "foo/13367",
      "_rev": "_cTyP57---A",
      "i": 10
    },
    {
      "_key": "13368",
      "_id": "foo/13368",
      "_rev": "_cTyP57---B",
      "i": 11
    },
    {
      "_key": "13372",
      "_id": "foo/13372",
      "_rev": "_cTyP57---C",
      "i": 15
    },
    {
      "_key": "13373",
      "_id": "foo/13373",
      "_rev": "_cTyP57C---",
      "i": 16
    },
    {
      "_key": "13374",
      "_id": "foo/13374",
      "_rev": "_cTyP57C--_",
      "i": 17
    },
    {
      "_key": "13376",
      "_id": "foo/13376",
      "_rev": "_cTyP57C--A",
      "i": 19
    },
    {
      "_key": "13377",
      "_id": "foo/13377",
      "_rev": "_cTyP57C--B",
      "i": 20
    },
    {
      "_key": "13382",
      "_id": "foo/13382",
      "_rev": "_cTyP57C--C",
      "i": 25
    }
  ],
  "hasMore": true,
  "cached": false,
  "extra": {
    "warnings": [],
    "stats": {
      "writesExecuted": 0,
      "writesIgnored": 0,
      "scannedFull": 20,
      "scannedIndex": 0,
      "filtered": 0,
      "httpRequests": 6,
      "executionTime": 0.003913449996616691,
      "peakMemoryUsage": 0
    }
  },
  "error": false,
  "code": 201
}`), &req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = false
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 404, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 120}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryPutNotFoundLongUptime(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	mockClient := util.NewMockClient(t, queryPutNotFoundLongUptime)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPutInternalError(
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
	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = false
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 500, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 120}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryPutInternalError(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	mockClient := util.NewMockClient(t, queryPutInternalError)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryHasMoreFalse(
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
	json.Unmarshal([]byte(`{
  "result": [
    {
      "_key": "13360",
      "_id": "foo/13360",
      "_rev": "_cTyP566---",
      "i": 3
    },
    {
      "_key": "13365",
      "_id": "foo/13365",
      "_rev": "_cTyP57----",
      "i": 8
    },
    {
      "_key": "13367",
      "_id": "foo/13367",
      "_rev": "_cTyP57---A",
      "i": 10
    },
    {
      "_key": "13368",
      "_id": "foo/13368",
      "_rev": "_cTyP57---B",
      "i": 11
    },
    {
      "_key": "13372",
      "_id": "foo/13372",
      "_rev": "_cTyP57---C",
      "i": 15
    },
    {
      "_key": "13373",
      "_id": "foo/13373",
      "_rev": "_cTyP57C---",
      "i": 16
    },
    {
      "_key": "13374",
      "_id": "foo/13374",
      "_rev": "_cTyP57C--_",
      "i": 17
    },
    {
      "_key": "13376",
      "_id": "foo/13376",
      "_rev": "_cTyP57C--A",
      "i": 19
    },
    {
      "_key": "13377",
      "_id": "foo/13377",
      "_rev": "_cTyP57C--B",
      "i": 20
    },
    {
      "_key": "13382",
      "_id": "foo/13382",
      "_rev": "_cTyP57C--C",
      "i": 25
    }
  ],
  "hasMore": false,
  "cached": false,
  "extra": {
    "warnings": [],
    "stats": {
      "writesExecuted": 0,
      "writesIgnored": 0,
      "scannedFull": 20,
      "scannedIndex": 0,
      "filtered": 0,
      "httpRequests": 6,
      "executionTime": 0.003913449996616691,
      "peakMemoryUsage": 0
    }
  },
  "error": false,
  "code": 201
}`), &req.Result)
	json.Unmarshal([]byte(`{"hasMore": true, "id": "id0", result": {"uptime": 120}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryHasMoreFalse(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	mockClient := util.NewMockClient(t, queryHasMoreFalse)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

func queryPutUnavailable(
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
	body := CursorResponse {
		HasMore: true,
		ID: "id0",
	}
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	body.HasMore = true
	*req.Result.(*CursorResponse) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 503, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 120}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// Get a normal PUT with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "PUT" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	path = "/_api/cursor/" + body.ID;
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	json.Unmarshal([]byte(`{
  "result": [
    {
      "_key": "13360",
      "_id": "foo/13360",
      "_rev": "_cTyP566---",
      "i": 3
    },
    {
      "_key": "13365",
      "_id": "foo/13365",
      "_rev": "_cTyP57----",
      "i": 8
    },
    {
      "_key": "13367",
      "_id": "foo/13367",
      "_rev": "_cTyP57---A",
      "i": 10
    },
    {
      "_key": "13368",
      "_id": "foo/13368",
      "_rev": "_cTyP57---B",
      "i": 11
    },
    {
      "_key": "13372",
      "_id": "foo/13372",
      "_rev": "_cTyP57---C",
      "i": 15
    },
    {
      "_key": "13373",
      "_id": "foo/13373",
      "_rev": "_cTyP57C---",
      "i": 16
    },
    {
      "_key": "13374",
      "_id": "foo/13374",
      "_rev": "_cTyP57C--_",
      "i": 17
    },
    {
      "_key": "13376",
      "_id": "foo/13376",
      "_rev": "_cTyP57C--A",
      "i": 19
    },
    {
      "_key": "13377",
      "_id": "foo/13377",
      "_rev": "_cTyP57C--B",
      "i": 20
    }
  ],
  "hasMore": false,
  "cached": false,
  "extra": {
    "warnings": [],
    "stats": {
      "writesExecuted": 0,
      "writesIgnored": 0,
      "scannedFull": 20,
      "scannedIndex": 0,
      "filtered": 0,
      "httpRequests": 6,
      "executionTime": 0.003913449996616691,
      "peakMemoryUsage": 0
    }
  },
  "error": false,
  "code": 201
}`), &req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
		Err:  nil,
	}

	// Get a normal Get with ID
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	path = "/_admin/statistics"
	if req.UrlPath != path {
		t.Errorf("Got wrong URL path %s instead of %s", req.UrlPath, path)
	}
	// Answer with a normal good response: created.
	json.Unmarshal([]byte(`{"server": {"uptime": 120}}`), req.Result)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse {
			StatusCode: 201, CoordinatorURL: "https://coordinator:8529" },
		Err:  nil,
	}

	// We have not responded with any count in the cursor yet, so not PUT is expected.
	next(ctx, t, requests, false)
}

func TestQueryPutUnavailable(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	mockClient := util.NewMockClient(t, queryPutUnavailable)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err != nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

