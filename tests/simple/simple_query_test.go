package simple

import (
	"context"
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

func queryDocumentsCursorUnavailable(
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

func TestQueryDocumentsCursorUnavailable(t *testing.T) {
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

	mockClient := util.NewMockClient(t, queryDocumentsCursorUnavailable)
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

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
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
		Resp: util.ArangoResponse { StatusCode: 201, },
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
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse { StatusCode: 200, },
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

	for i := 0; i < 10; i++ {
		doc := UserDocument{
			Key:   foo.createNewKey(true),
			Value: rand.Int(),
			Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
			Odd:   time.Now().Nanosecond()%2 == 1,
		}
		foo.existingDocs[doc.Key] = doc
	}

	mockClient := util.NewMockClient(t, queryDocumentsCursorHasMore)
	test.client = mockClient
	test.listener = util.MockListener{}

	// Expecting error, as count from cursor does not match.
	err := test.queryDocuments(foo)
	if err == nil {
		t.Errorf("Unexpected result from queryDocuments: err: %v", err)
	}
	mockClient.Shutdown()
}

