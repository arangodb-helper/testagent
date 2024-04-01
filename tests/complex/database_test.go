package complex

import (
	"context"
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

func createDatabaseOkBehaviour(
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
	expectedName := "test_onehard_db"
	expectedSharding := "single"
	expectedReplicationFactor := 2
	expectedWriteConcern := 1

	//validate values in actual requets
	actualRequestBody := req.Input.(map[string]interface{})
	//database name
	{
		actualName, ok := actualRequestBody["name"].(string)
		if !ok {
			t.Errorf("Request body does not contain an expected field \"name\"")
		}
		if actualName != expectedName {
			t.Errorf("Wrong value in request field \"name\". Expected: %s, actual: %s", expectedName, actualName)
		}
	}
	//sharding
	{
		actualSharding, ok := actualRequestBody["options"].(map[string]interface{})["sharding"]
		if !ok {
			t.Errorf("Request body does not contain an expected field \"options.sharding\"")
		}
		if actualSharding != expectedSharding {
			t.Errorf("Wrong value in request field \"options.sharding\". Expected: %s, actual: %s", expectedName, actualSharding)
		}
	}
	//replicationFactor
	{
		actualReplicationFactor, ok := actualRequestBody["options"].(map[string]interface{})["replicationFactor"]
		if !ok {
			t.Errorf("Request body does not contain an expected field \"options.replicationFactor\"")
		}
		if actualReplicationFactor != expectedReplicationFactor {
			t.Errorf("Wrong value in request field \"options.replicationFactor\". Expected: %d, actual: %d", expectedReplicationFactor, actualReplicationFactor)
		}
	}
	//writeConcern
	{
		actualWriteConcern, ok := actualRequestBody["options"].(map[string]interface{})["writeConcern"]
		if !ok {
			t.Errorf("Request body does not contain an expected field \"options.writeConcern\"")
		}
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

func TestCreateDatabase(t *testing.T) {
	test := &OneShardTest{
		DocColTest: DocColTest{
			ComplextTest: ComplextTest{
				TestName: "MockOneShardDbTest",
				ComplextTestContext: ComplextTestContext{
					ComplextTestConfig: ComplextTestConfig{
						NumberOfShards:    1,
						ReplicationFactor: 2,
						OperationTimeout:  time.Second * 10,
						RetryTimeout:      time.Second * 10,
						StepTimeout:       time.Second,
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
	mockClient := util.NewMockClient(t, createDatabaseOkBehaviour)
	test.client = mockClient
	test.listener = util.MockListener{}
	err := test.createOneShardDatabase("test_onehard_db")
	if err != nil {
		t.Errorf("Unexpected error while creating a oneshard DB: %v", err)
	}
}
