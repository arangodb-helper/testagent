package simple

import (
	"context"
	"errors"
	"testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

func importDocumentsOkBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/import" {
		t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
	}

	// Answer with a normal good response:
	body := map[string]interface{}{
		"created": float64(10000),
		"details": []int{1, 2, 3},
	}
	*req.Result.(*interface{}) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestImportDocumentsOk(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, importDocumentsOkBehaviour)
	test.client = mockClient
	test.listener = util.MockListener{}
	err := test.importDocuments(coll)
	if err != nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	mockClient.Shutdown()
}

func importDocumentsErrorBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/import" {
		t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
	}

	// Answer with an error response:
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  errors.New("Unexpected code"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestImportDocumentsError(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, importDocumentsErrorBehaviour)
	test.client = mockClient
	test.listener = util.MockListener{}
	err := test.importDocuments(coll)
	if err == nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	mockClient.Shutdown()
}

func importDocumentsIncompleteBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/import" {
		t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
	}

	// Answer with a normal good response:
	body := map[string]interface{}{
		"created": float64(8000),
		"details": []int{1, 2, 3},
	}
	*req.Result.(*interface{}) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// Now the same but without details:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/import" {
		t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
	}

	// Answer with a normal good response:
	body = map[string]interface{}{
		"created": float64(8000),
	}
	*req.Result.(*interface{}) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestImportDocumentsIncomplete(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, importDocumentsIncompleteBehaviour)
	test.client = mockClient
	test.listener = util.MockListener{}
	err := test.importDocuments(coll)
	if err == nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	err = test.importDocuments(coll)
	if err == nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	mockClient.Shutdown()
}

func importDocumentsBadResponseBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a normal POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/import" {
		t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
	}

	// Answer with an bad value for created:
	body := map[string]interface{}{
		"created": "bla",
		"details": []int{1, 2, 3},
	}
	*req.Result.(*interface{}) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// Now without value for created:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/import" {
		t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
	}

	// Answer with a normal good response:
	body = map[string]interface{}{
		"noCreatedHere": "bla",
	}
	*req.Result.(*interface{}) = body
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// Now with a nonsensical body:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	if req.UrlPath != "/_api/import" {
		t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
	}

	// Answer with a normal good response:
	body2 := "nonsense"
	*req.Result.(*interface{}) = body2
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestImportDocumentsBadResponse(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, importDocumentsBadResponseBehaviour)
	test.client = mockClient
	test.listener = util.MockListener{}
	err := test.importDocuments(coll)
	if err == nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	err = test.importDocuments(coll)
	if err == nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	err = test.importDocuments(coll)
	if err == nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	mockClient.Shutdown()
}

func importDocumentsTimeoutBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

  for {
		// Get a normal POST request:
		req := potentialNext(ctx, t, requests)
		if req == nil {
			return
		}
		if req.Method != "POST" {
			t.Errorf("Got wrong method %s instead of POST.", req.Method)
		}
		if req.UrlPath != "/_api/import" {
			t.Errorf("Got wrong URL path %s instead of /_api/import", req.UrlPath)
		}

		// Answer with an error response:
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 0},
			Err:  nil,
		}
	}
}

func TestImportDocumentsTimeout(t *testing.T) {
	test := simpleTest{
		SimpleConfig: config,
		reportDir:    ".",
		log:          log,
		collections:  make(map[string]*collection),
	}
	mockClient := util.NewMockClient(t, importDocumentsTimeoutBehaviour)
	test.client = mockClient
	test.listener = util.MockListener{}
	err := test.importDocuments(coll)
	if err == nil {
		t.Errorf("Unexpected result from importDocuments, err: %v", err)
	}
	mockClient.Shutdown()
}

