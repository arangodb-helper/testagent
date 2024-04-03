package complex

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb-helper/testagent/tests/util"
	logging "github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("testAgentTests")
)

func next(ctx context.Context, t *testing.T, requests chan *util.MockRequest, expectMore bool) *util.MockRequest {
	select {
	case req := <-requests:
		if !expectMore {
			t.Errorf("Did not expect further request, got: %v.", req)
		}
		return req
	case <-ctx.Done():
		if expectMore {
			t.Errorf("Expecting further requests.")
		}
		return nil
	}
}

func NewMockTest(mockClient *util.MockClient) *OneShardTest {
	test := &OneShardTest{
		DocColTest: DocColTest{
			ComplextTest: ComplextTest{
				TestName: "MockOneShardDbTest",
				ComplextTestContext: ComplextTestContext{
					ComplextTestConfig: ComplextTestConfig{
						NumberOfShards:    1,
						ReplicationFactor: 2,
						OperationTimeout:  time.Millisecond * 300,
						RetryTimeout:      time.Millisecond * 300,
						StepTimeout:       time.Millisecond * 100,
					},
					ComplextTestHarness: ComplextTestHarness{
						reportDir: ".",
						log:       log,
					},
					documentIdSeq:     0,
					collectionNameSeq: 0,
					existingDocuments: make([]TestDocument, 0, 1000),
				},
			},
			DocColConfig:             DocColConfig{},
			numberOfExistingDocs:     0,
			numberOfCreatedDocsTotal: 0,
			docCollectionCreated:     false,
		},
		databaseNameSeq:   0,
		isDatabaseCreated: false,
	}
	test.DocColTestImpl = test
	test.ComplexTestImpl = test
	test.client = mockClient
	test.listener = util.MockListener{}
	return test
}

func createOneShardDatabaseOkBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}
	//check body
	// extected values:
	expectedName := "test_oneshard_db"
	expectedSharding := "single"
	expectedReplicationFactor := 2
	expectedWriteConcern := 1

	//validate values in actual requets
	actualRequestBody := req.Input.(CreateDatabaseRequest)
	//database name
	{
		actualName := actualRequestBody.Name
		if actualName != expectedName {
			t.Errorf("Wrong value in request field \"name\". Expected: %s, actual: %s", expectedName, actualName)
		}
	}
	//sharding
	{
		actualSharding := actualRequestBody.Options.Sharding
		if actualSharding != expectedSharding {
			t.Errorf("Wrong value in request field \"options.sharding\". Expected: %s, actual: %s", expectedName, actualSharding)
		}
	}
	//replicationFactor
	{
		actualReplicationFactor := actualRequestBody.Options.ReplicationFactor
		if actualReplicationFactor != expectedReplicationFactor {
			t.Errorf("Wrong value in request field \"options.replicationFactor\". Expected: %d, actual: %d", expectedReplicationFactor, actualReplicationFactor)
		}
	}
	//writeConcern
	{
		actualWriteConcern := actualRequestBody.Options.WriteConcern
		if actualWriteConcern != expectedWriteConcern {
			t.Errorf("Wrong value in request field \"options.writeConcern\". Expected: %d, actual: %d", expectedWriteConcern, actualWriteConcern)
		}
	}
	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateOneShardDatabase(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createOneShardDatabaseOkBehaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err != nil {
		t.Errorf("Unexpected error while creating a oneshard DB: %v", err)
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.succeeded != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createOneShardDatabaseWithRecheckBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	//now the client must check if the database was created

	// Get a GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	req.Result.(*DatabasesResponse).Result = []string{"_system", "test_oneshard_db"}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateOneShardDatabaseWithRecheck(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createOneShardDatabaseWithRecheckBehaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err != nil {
		t.Errorf("Unexpected error while creating a oneshard DB: %v", err)
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.succeeded != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createOneShardDatabaseWithRetry500Behaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//send a 500 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 500},
		Err:  nil,
	}

	//now the client must check if the database exists. we shall pretend it doesn't

	// Get a GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	req.Result.(*DatabasesResponse).Result = []string{"_system"}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	//now the client must attempt to create the database again
	// Get a POST request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}
	//check body
	// extected values:
	expectedName := "test_oneshard_db"
	expectedSharding := "single"
	expectedReplicationFactor := 2
	expectedWriteConcern := 1

	//validate values in actual requets
	actualRequestBody := req.Input.(CreateDatabaseRequest)
	//database name
	{
		actualName := actualRequestBody.Name
		if actualName != expectedName {
			t.Errorf("Wrong value in request field \"name\". Expected: %s, actual: %s", expectedName, actualName)
		}
	}
	//sharding
	{
		actualSharding := actualRequestBody.Options.Sharding
		if actualSharding != expectedSharding {
			t.Errorf("Wrong value in request field \"options.sharding\". Expected: %s, actual: %s", expectedName, actualSharding)
		}
	}
	//replicationFactor
	{
		actualReplicationFactor := actualRequestBody.Options.ReplicationFactor
		if actualReplicationFactor != expectedReplicationFactor {
			t.Errorf("Wrong value in request field \"options.replicationFactor\". Expected: %d, actual: %d", expectedReplicationFactor, actualReplicationFactor)
		}
	}
	//writeConcern
	{
		actualWriteConcern := actualRequestBody.Options.WriteConcern
		if actualWriteConcern != expectedWriteConcern {
			t.Errorf("Wrong value in request field \"options.writeConcern\". Expected: %d, actual: %d", expectedWriteConcern, actualWriteConcern)
		}
	}
	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateOneShardDatabaseWithRetry500(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createOneShardDatabaseWithRetry500Behaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err != nil {
		t.Errorf("Unexpected error while creating a oneshard DB: %v", err)
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.succeeded != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createOneShardDatabaseWithRetry503Behaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	//now the client must check if the database exists. we shall pretend it doesn't

	// Get a GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	req.Result.(*DatabasesResponse).Result = []string{"_system"}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	//now the client must attempt to create the database again
	// Get a POST request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}
	//check body
	// extected values:
	expectedName := "test_oneshard_db"
	expectedSharding := "single"
	expectedReplicationFactor := 2
	expectedWriteConcern := 1

	//validate values in actual requets
	actualRequestBody := req.Input.(CreateDatabaseRequest)
	//database name
	{
		actualName := actualRequestBody.Name
		if actualName != expectedName {
			t.Errorf("Wrong value in request field \"name\". Expected: %s, actual: %s", expectedName, actualName)
		}
	}
	//sharding
	{
		actualSharding := actualRequestBody.Options.Sharding
		if actualSharding != expectedSharding {
			t.Errorf("Wrong value in request field \"options.sharding\". Expected: %s, actual: %s", expectedName, actualSharding)
		}
	}
	//replicationFactor
	{
		actualReplicationFactor := actualRequestBody.Options.ReplicationFactor
		if actualReplicationFactor != expectedReplicationFactor {
			t.Errorf("Wrong value in request field \"options.replicationFactor\". Expected: %d, actual: %d", expectedReplicationFactor, actualReplicationFactor)
		}
	}
	//writeConcern
	{
		actualWriteConcern := actualRequestBody.Options.WriteConcern
		if actualWriteConcern != expectedWriteConcern {
			t.Errorf("Wrong value in request field \"options.writeConcern\". Expected: %d, actual: %d", expectedWriteConcern, actualWriteConcern)
		}
	}
	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 201},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateOneShardDatabaseWithRetry503(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createOneShardDatabaseWithRetry503Behaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err != nil {
		t.Errorf("Unexpected error while creating a oneshard DB: %v", err)
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.succeeded != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createOneShardDatabaseButItExistsBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//send a 409 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateOneShardDatabaseButItExists(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createOneShardDatabaseButItExistsBehaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err == nil {
		t.Errorf("unexpected result from createOneShardDatabase: must return an error")
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.failed != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createOneShardDatabaseButItExistsWithRetryBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//send a 500 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 500},
		Err:  nil,
	}

	//now the client must check if the database exists. we shall pretend that it does

	// Get a GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	req.Result.(*DatabasesResponse).Result = []string{"_system", "test_oneshard_db"}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateOneShardDatabaseButItExistsWithRetry(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createOneShardDatabaseButItExistsWithRetryBehaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err == nil {
		t.Errorf("unexpected result from createOneShardDatabase: must return an error")
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.failed != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createOneShardDatabaseFirstAttemptUncertainButSuccessfullBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a POST request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	//now the client must check if the database exists. we shall pretend that it doesn't

	// Get a GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	req.Result.(*DatabasesResponse).Result = []string{"_system"}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// Get a POST request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	//send a 409 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}

	//now the client must check if the database exists(again). this time we shall pretend that it does

	// Get a GET request:
	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	if req.UrlPath != "/_api/database" {
		t.Errorf("Got wrong URL path %s instead of /_api/database",
			req.UrlPath)
	}

	req.Result.(*DatabasesResponse).Result = []string{"_system", "test_oneshard_db"}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCreateOneShardDatabaseFirstAttemptUncertainButSuccessfull(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createOneShardDatabaseFirstAttemptUncertainButSuccessfullBehaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err != nil {
		t.Errorf("Unexpected error while creating a oneshard DB: %v", err)
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.succeeded != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createDatabaseTestTimeOutBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	for {
		req := next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method != "GET" && req.UrlPath != "/_api/database" {
			//send a successfull response
			req.Result.(*DatabasesResponse).Result = []string{"_system"}
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 200},
				Err:  nil,
			}
		} else {
			//send a 503 response
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 503},
				Err:  nil,
			}
		}
	}
}

func TestCreateDatabaseTestTimeOut(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createDatabaseTestTimeOutBehaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err == nil {
		t.Errorf("unexpected result from createOneShardDatabase: must return an error")
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.failed != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the an attempt to create a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.succeeded)
	}
}

func createDatabaseErrorBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	for {
		req := next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method == "GET" && req.UrlPath != "/_api/database" {
			//send a successfull response
			req.Result.(*DatabasesResponse).Result = []string{"_system"}
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 200},
				Err:  nil,
			}
		} else {
			//return an error
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 500},
				Err:  fmt.Errorf("Error"),
			}
		}
	}
}

func TestCreateDatabaseError(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, createDatabaseErrorBehaviour))
	err := test.createOneShardDatabase("test_oneshard_db")
	if err == nil {
		t.Errorf("an error was expected, but the function did not return it")
	}
	expectedCounterValue := 1
	if test.createDatabaseCounter.failed != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the failed attempt to drop a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.createDatabaseCounter.failed)
	}
}

func dropDatabaseOkBehaviour(
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
	if req.UrlPath != "/_api/database/database_to_drop_name" {
		t.Errorf("Got wrong URL path %s instead of /_api/database/database_to_drop_name",
			req.UrlPath)
	}

	//send a successfull response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestDropDatabase(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, dropDatabaseOkBehaviour))
	err := test.dropDatabase("database_to_drop_name")
	if err != nil {
		t.Errorf("Unexpected error while dropping a DB: %v", err)
	}
	expectedCounterValue := 1
	if test.dropDatabaseCounter.succeeded != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the failed attempt to drop a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.dropDatabaseCounter.succeeded)
	}
}

func dropDatabaseWithNetworkProblemsBehaviour(
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
	if req.UrlPath != "/_api/database/database_to_drop_name" {
		t.Errorf("Got wrong URL path %s instead of /_api/database/database_to_drop_name",
			req.UrlPath)
	}

	//send a 503 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	//now we expect the client to retry
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
	if req.UrlPath != "/_api/database/database_to_drop_name" {
		t.Errorf("Got wrong URL path %s instead of /_api/database/database_to_drop_name",
			req.UrlPath)
	}

	//this time we send a 404 response, as if the database was dropped after the first request
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestDropDatabaseWithProblems(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, dropDatabaseWithNetworkProblemsBehaviour))
	err := test.dropDatabase("database_to_drop_name")
	if err != nil {
		t.Errorf("Unexpected error while dropping a DB: %v", err)
	}
	expectedCounterValue := 1
	if test.dropDatabaseCounter.succeeded != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the failed attempt to drop a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.dropDatabaseCounter.succeeded)
	}
}

func dropNonExistingDatabaseBehaviour(
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
	if req.UrlPath != "/_api/database/database_to_drop_name" {
		t.Errorf("Got wrong URL path %s instead of /_api/database/database_to_drop_name",
			req.UrlPath)
	}

	//send a 404 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestDropNonExistingDatabase(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, dropNonExistingDatabaseBehaviour))
	err := test.dropDatabase("database_to_drop_name")
	if err == nil {
		t.Errorf("an error was expected, but the function did not return it")
	}
	expectedCounterValue := 1
	if test.dropDatabaseCounter.failed != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the failed attempt to drop a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.dropDatabaseCounter.failed)
	}
}

func dropDatabaseTestTimeOutBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	for {
		req := next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method == "GET" && req.UrlPath != "/_api/database" {
			//send a successfull response
			req.Result.(*DatabasesResponse).Result = []string{"_system", "database_to_drop_name"}
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 200},
				Err:  nil,
			}
		} else {
			//send a 503 response
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 503},
				Err:  nil,
			}
		}
	}
}

func TestDropDatabaseTimeout(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, dropDatabaseTestTimeOutBehaviour))
	err := test.dropDatabase("database_to_drop_name")
	if err == nil {
		t.Errorf("an error was expected, but the function did not return it")
	}
	expectedCounterValue := 1
	if test.dropDatabaseCounter.failed != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the failed attempt to drop a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.dropDatabaseCounter.failed)
	}
}

func dropDatabaseErrorBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	for {
		req := next(ctx, t, requests, true)
		if req == nil {
			return
		}
		if req.Method == "GET" && req.UrlPath != "/_api/database" {
			//send a successfull response
			req.Result.(*DatabasesResponse).Result = []string{"_system", "database_to_drop_name"}
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 200},
				Err:  nil,
			}
		} else {
			//return an error
			responses <- &util.MockResponse{
				Resp: util.ArangoResponse{StatusCode: 500},
				Err:  fmt.Errorf("Error"),
			}
		}
	}
}

func TestDropDatabaseError(t *testing.T) {
	test := NewMockTest(util.NewMockClient(t, dropDatabaseErrorBehaviour))
	err := test.dropDatabase("database_to_drop_name")
	if err == nil {
		t.Errorf("an error was expected, but the function did not return it")
	}
	expectedCounterValue := 1
	if test.dropDatabaseCounter.failed != expectedCounterValue {
		t.Errorf("counter value wasn't raised after the failed attempt to drop a db. expected value: %d, actual value: %d",
			expectedCounterValue, test.dropDatabaseCounter.failed)
	}
}
