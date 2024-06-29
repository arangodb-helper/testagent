package complex

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb-helper/testagent/tests/util"
)

var (
	graphName = "graph_name"
)

func NewMockGraphTest(mockClient *util.MockClient) *EnterpriseGraphTest {
	test := &EnterpriseGraphTest{
		SmartGraphTest: SmartGraphTest{
			GraphTest: GraphTest{
				ComplextTest: ComplextTest{
					TestName: "MockGraphTest",
					ComplextTestContext: ComplextTestContext{
						ComplextTestConfig: ComplextTestConfig{
							NumberOfShards:    1,
							ReplicationFactor: 2,
							OperationTimeout:  time.Millisecond * 100,
							RetryTimeout:      time.Millisecond * 100,
							StepTimeout:       time.Millisecond * 5,
						},
						ComplextTestHarness: ComplextTestHarness{
							reportDir: ".",
							log:       log,
						},
						documentIdSeq:     0,
						collectionNameSeq: 0,
						existingDocuments: make([]TestDocument, 0, 1000),
					},
				}}},
	}
	test.GraphTestImpl = test
	test.ComplexTestImpl = test
	test.client = mockClient
	test.listener = util.MockListener{}
	return test
}

// a template for testing the createGraph function
func checkCreateGraph(t *testing.T, expectError bool, behaviour util.Behaviour) {
	test := NewMockGraphTest(util.NewMockClient(t, behaviour))
	err := test.createGraph(graphName, "edge_col", []string{"from_col"}, []string{"to_col"}, []string{}, true, false, "smartattr", []string{}, 100, 3, 2)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected result from checkIfDocumentExists: must return an error")
	}
}

func createGraphOKBehaviour(
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
	expected_url := "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraphOKTest(t *testing.T) {
	checkCreateGraph(t, false, createGraphOKBehaviour)
}

func createGraph503Behaviour(
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
	expected_url := "/_api/gharial"
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
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
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

func TestCreateGraph503Test(t *testing.T) {
	checkCreateGraph(t, false, createGraph503Behaviour)
}

func createGraphRetryBehaviour(
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
	expected_url := "/_api/gharial"
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
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 404 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

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
	expected_url = "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response this time
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraphRetryTest(t *testing.T) {
	checkCreateGraph(t, false, createGraphRetryBehaviour)
}

func createGraph500RetryBehaviour(
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
	expected_url := "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 500 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 500},
		Err:  nil,
	}

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

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 404 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

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
	expected_url = "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response this time
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraph500RetryTest(t *testing.T) {
	checkCreateGraph(t, false, createGraph500RetryBehaviour)
}

func createGraph500ErrorBehaviour(
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
	expected_url := "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 500 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 500},
		Err:  nil,
	}

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

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 200 response, as if the graph exists
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraph500ErrorTest(t *testing.T) {
	checkCreateGraph(t, true, createGraph500ErrorBehaviour)
}

func createGraphConflictBehaviour(
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
	expected_url := "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 409 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraphConflictTest(t *testing.T) {
	checkCreateGraph(t, true, createGraphConflictBehaviour)
}

func createGraph503ThenConflictBehaviour(
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
	expected_url := "/_api/gharial"
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
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 404 response, as if the graph exists
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

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
	expected_url = "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 409 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}

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

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
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

func TestCreateGraph503ThenConflictTest(t *testing.T) {
	checkCreateGraph(t, false, createGraph503ThenConflictBehaviour)
}

func createGraph503Then404LoopBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	for {
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
		expected_url := "/_api/gharial"
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
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}
		//check URL

		expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
		if req.UrlPath != expected_url {
			t.Errorf("Got wrong URL path %s instead of %s",
				req.UrlPath, expected_url)
		}

		//send a 404 response, as if the graph exists
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 404},
			Err:  nil,
		}
	}
}

func TestCreateGraphTimeoutTest(t *testing.T) {
	savedReadTimeout := ReadTimeout
	ReadTimeout = readTimeoutForTesting // to speed up tests
	defer func() { ReadTimeout = savedReadTimeout }()
	checkCreateGraph(t, true, createGraph503Then404LoopBehaviour)
}

func createGraphErrorBehaviour(
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
	expected_url := "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send an error response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  fmt.Errorf("Error"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraphErrorTest(t *testing.T) {
	checkCreateGraph(t, true, createGraphErrorBehaviour)
}

func createGraphCheckErrorBehaviour(
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
	expected_url := "/_api/gharial"
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
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send an error response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 0},
		Err:  fmt.Errorf(""),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraphCheckErrorTest(t *testing.T) {
	checkCreateGraph(t, true, createGraphCheckErrorBehaviour)
}

func createGraphButItIsBrokenBehaviour(
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
	expected_url := "/_api/gharial"
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
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send an error response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

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
	expected_url = "/_api/gharial"
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 409 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}

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

	expected_url = fmt.Sprintf("/_api/gharial/%s", graphName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send an error response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateGraphButItIsBrokenTest(t *testing.T) {
	checkCreateGraph(t, true, createGraphButItIsBrokenBehaviour)
}
