package complex

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb-helper/testagent/tests/util"
)

var (
	graphName     = "graph_name"
	edgeColName   = "edge_col"
	vertexColName = "vertex_col"
	to            = "0002"
	from          = "0001"
)

func NewMockGraphTest(mockClient *util.MockClient) *EnterpriseGraphTest {
	test := &EnterpriseGraphTest{
		SmartGraphTest: SmartGraphTest{
			GraphTest: GraphTest{
				GraphTestConf: GraphTestConf{
					VertexSize: 4,
					EdgeSize:   4,
				},
				ComplextTest: ComplextTest{
					TestName: "MockGraphTest",
					ComplextTestContext: ComplextTestContext{
						ComplextTestConfig: ComplextTestConfig{
							NumberOfShards:    1,
							ReplicationFactor: 2,
							// OperationTimeout:  time.Second * 90,
							OperationTimeout: time.Millisecond * 500,
							RetryTimeout:     time.Millisecond * 100,
							StepTimeout:      time.Millisecond * 5,
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
		t.Errorf("unexpected result from createGraph: must return an error")
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

// a template for testing the createEdge function
func checkCreateEdge(t *testing.T, expectError bool, behaviour util.Behaviour) {
	test := NewMockGraphTest(util.NewMockClient(t, behaviour))
	err := test.createEdge(to, from, edgeColName, vertexColName, 4)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected result from createEdge: must return an error")
	}
}

func createEdgeOKBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
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

func TestCreateEdgeOKTest(t *testing.T) {
	checkCreateEdge(t, false, createEdgeOKBehaviour)
}

func createEdge503SuccessBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	// Expect a request to read the created document
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	expected_url = "/_api/cursor"
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	// set request result
	reqCursorResponse, _ := req.Result.(*CursorResponse)
	reqCursorResponse.Result = []interface{}{map[string]interface{}{
		"_key": "key", "_rev": "rev", "seed": 1.0, "update_counter": 1.0,
	}}
	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdge503SuccessTest(t *testing.T) {
	checkCreateEdge(t, false, createEdge503SuccessBehaviour)
}

func createEdge503RetryBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	// Expect a request to read the created document
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	expected_url = "/_api/cursor"
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response but leave cursor empty
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
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
	expected_url = fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdge503RetryTest(t *testing.T) {
	checkCreateEdge(t, false, createEdge503RetryBehaviour)
}

func createEdgeErrorBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//return an error
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  fmt.Errorf("Error"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdgeErrorTest(t *testing.T) {
	checkCreateEdge(t, true, createEdgeErrorBehaviour)
}

func createEdgeInvalidResponseBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//check query param
	value, exists := req.Query["waitForSync"]
	if !(exists && (len(value) == 1 && value[0] == "true")) {
		t.Errorf("\"waitForSync\" query param must be set to \"true\"")
	}

	//return an invalid status code
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 202},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdgeInvalidResponseTest(t *testing.T) {
	checkCreateEdge(t, true, createEdgeInvalidResponseBehaviour)
}

func createEdge410ButCreatedResponseBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//check query param
	value, exists := req.Query["waitForSync"]
	if !(exists && (len(value) == 1 && value[0] == "true")) {
		t.Errorf("\"waitForSync\" query param must be set to \"true\"")
	}

	//return 410
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}

	// Expect a request to read the created document
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	expected_url = "/_api/cursor"
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response
	req.Result.(*CursorResponse).Result = []interface{}{
		map[string]interface{}{"seed": 1.0, "_key": "1", "_rev": "abcxyz", "update_counter": 0.0}}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdge410ResponseTest(t *testing.T) {
	checkCreateEdge(t, true, createEdge410ButCreatedResponseBehaviour)
}

func createEdge410RetryBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//check query param
	value, exists := req.Query["waitForSync"]
	if !(exists && (len(value) == 1 && value[0] == "true")) {
		t.Errorf("\"waitForSync\" query param must be set to \"true\"")
	}

	//return 410
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 410},
		Err:  nil,
	}

	// Expect a request to read the created document
	// Get a request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	expected_url = "/_api/cursor"
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a successfull response with empty cursor
	*&req.Result.(*CursorResponse).Result = []interface{}{}
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
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
	expected_url = fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//check query param
	value, exists = req.Query["waitForSync"]
	if !(exists && (len(value) == 1 && value[0] == "true")) {
		t.Errorf("\"waitForSync\" query param must be set to \"true\"")
	}

	//return successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdge410RetryTest(t *testing.T) {
	checkCreateEdge(t, false, createEdge410RetryBehaviour)
}

func createEdgeTransactionLostRetryBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//check query param
	value, exists := req.Query["waitForSync"]
	if !(exists && (len(value) == 1 && value[0] == "true")) {
		t.Errorf("\"waitForSync\" query param must be set to \"true\"")
	}

	//return 404 with arango error 1655
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404, Error_: util.ArangoError{Error_: true, ErrorMessage: "transaction not found", Code: 404, ErrorNum: 1655}},
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
	expected_url = fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//check query param
	value, exists = req.Query["waitForSync"]
	if !(exists && (len(value) == 1 && value[0] == "true")) {
		t.Errorf("\"waitForSync\" query param must be set to \"true\"")
	}

	//return successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdgeTransactionLostRetryTest(t *testing.T) {
	checkCreateEdge(t, false, createEdgeTransactionLostRetryBehaviour)
}

func createEdge404FailBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", edgeColName)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//check query param
	value, exists := req.Query["waitForSync"]
	if !(exists && (len(value) == 1 && value[0] == "true")) {
		t.Errorf("\"waitForSync\" query param must be set to \"true\"")
	}

	//return 404 with any arango error other than 1655(transaction lost)
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404, Error_: util.ArangoError{Error_: true, ErrorMessage: "error", Code: 404, ErrorNum: 1000}},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateEdge404FailTest(t *testing.T) {
	checkCreateEdge(t, true, createEdge404FailBehaviour)
}
