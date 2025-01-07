package complex

import (
	"context"
	"fmt"
	"testing"

	"github.com/arangodb-helper/testagent/tests/util"
)

// createCollection test template
func checkCreateCollection(t *testing.T, expectError bool, edgeCol bool, behaviour util.Behaviour) {
	savedBackOffTime := BackOffTime
	BackOffTime = backOffTimeForTesting // to speed up tests
	defer func() { BackOffTime = savedBackOffTime }()
	test := NewMockTest(util.NewMockClient(t, behaviour))
	err := test.createCollection(CollectionName, edgeCol)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected behaviour from createCollection: must return an error")
	}
}

// dropCollection test template
func checkDropCollection(t *testing.T, expectError bool, behaviour util.Behaviour) {
	savedBackOffTime := BackOffTime
	BackOffTime = backOffTimeForTesting // to speed up tests
	defer func() { BackOffTime = savedBackOffTime }()
	test := NewMockTest(util.NewMockClient(t, behaviour))
	err := test.dropCollection(CollectionName)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected behaviour from createCollection: must return an error")
	}
}

func createCollectionSuccessBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/collection")
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateDocColSuccess(t *testing.T) {
	checkCreateCollection(t, false, false, createCollectionSuccessBehaviour)
}

func TestCreateEdgeColSuccess(t *testing.T) {
	checkCreateCollection(t, false, true, createCollectionSuccessBehaviour)
}

func createCollectionRetryBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/collection")
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 410 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}

	//expect a GET request
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url = fmt.Sprintf("/_api/collection/%s", CollectionName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}
	//send a 404 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	//expect another attempt to create a collection via POST request
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	expected_url = fmt.Sprintf("/_api/collection")
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}
	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	//expect a GET request
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url = fmt.Sprintf("/_api/collection/%s", CollectionName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}
	//send a 200 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateDocColRetry(t *testing.T) {
	checkCreateCollection(t, false, false, createCollectionRetryBehaviour)
}

func TestCreateEdgeColRetry(t *testing.T) {
	checkCreateCollection(t, false, true, createCollectionRetryBehaviour)
}

func createCollection410ButExistsBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/collection")
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 410 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}

	//expect a GET request
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url = fmt.Sprintf("/_api/collection/%s", CollectionName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}
	//send a 200 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}
	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateDocCol410ButCreated(t *testing.T) {
	checkCreateCollection(t, true, false, createCollection410ButExistsBehaviour)
}

func TestCreateEdgeCol410ButCreated(t *testing.T) {
	checkCreateCollection(t, true, true, createCollection410ButExistsBehaviour)
}

func dropCollectionSuccessBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/collection/%s", CollectionName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestDropColSuccess(t *testing.T) {
	checkDropCollection(t, false, dropCollectionSuccessBehaviour)
}

func dropCollectionRetryBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/collection/%s", CollectionName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "DELETE" {
		t.Errorf("Got wrong method %s instead of DELETE.", req.Method)
	}
	//check URL
	expected_url = fmt.Sprintf("/_api/collection/%s", CollectionName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 200 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestDropColRetry(t *testing.T) {
	checkDropCollection(t, false, dropCollectionRetryBehaviour)
}
