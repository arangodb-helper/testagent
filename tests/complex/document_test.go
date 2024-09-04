package complex

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb-helper/testagent/tests/util"
)

var (
	CollectionName   string       = "some_collection"
	DocumentKey      string       = "some_key"
	ExpectedDocument BigDocument  = NewBigDocumentFromSeed(2, 16)
	WrongDocument    BigDocument  = NewBigDocumentFromSeed(3, 16)
	ChangedDocument  TestDocument = TestDocument{
		Seed:          ExpectedDocument.Seed,
		Key:           ExpectedDocument.Key,
		Rev:           "new_rev",
		UpdateCounter: ExpectedDocument.UpdateCounter + 1,
	}
	backOffTimeForTesting time.Duration = time.Millisecond * 2
	readTimeoutForTesting int           = 1
)

//checkIfDocumentExists tests

// a template for testing the checkIfDocumentExists function
func checkThatDocumentExists(t *testing.T, expectedResult bool, expectError bool, behaviour util.Behaviour) {
	savedBackOffTime := BackOffTime
	BackOffTime = backOffTimeForTesting // to speed up tests
	defer func() { BackOffTime = savedBackOffTime }()
	test := NewMockTest(util.NewMockClient(t, behaviour))
	actualResult, err := test.checkIfDocumentExists(CollectionName, DocumentKey)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected result from checkIfDocumentExists: must return an error")
	}
	if !expectError {
		if actualResult != expectedResult {
			t.Errorf("wrong value returned. expected: %t, actual value: %t",
				expectedResult, actualResult)
		}
	}
}

func checkIfDocumentExistsYesBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, DocumentKey)
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

func TestCheckIfDocumentExistsYes(t *testing.T) {
	checkThatDocumentExists(t, true, false, checkIfDocumentExistsYesBehaviour)
}

func checkIfDocumentExistsNoBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, DocumentKey)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//send a 404 response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestCheckIfDocumentExistsNo(t *testing.T) {
	checkThatDocumentExists(t, false, false, checkIfDocumentExistsNoBehaviour)
}

func checkIfDocumentExistsRetryBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
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
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, DocumentKey)
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

func TestCheckIfDocumentExistsRetry(t *testing.T) {
	checkThatDocumentExists(t, true, false, checkIfDocumentExistsRetryBehaviour)
}

func only503ReplyBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	for {
		// Get a request:
		req := next(ctx, t, requests, true)
		if req == nil {
			return
		}
		//send a 503 response
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestCheckIfDocumentExistsTimeOut(t *testing.T) {
	checkThatDocumentExists(t, true, true, only503ReplyBehaviour)
}

// insertDocument tests
// a template for testing the insertDocument function
func checkInsertDocument(t *testing.T, expectError bool, behaviour util.Behaviour) {
	savedBackOffTime := BackOffTime
	BackOffTime = backOffTimeForTesting // to speed up tests
	defer func() { BackOffTime = savedBackOffTime }()
	test := NewMockTest(util.NewMockClient(t, behaviour))
	document := NewBigDocumentFromSeed(1, 16)
	err := test.insertDocument(CollectionName, document)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected result from insertDocument: must return an error")
	}
	if expectError {
		if test.singleDocCreateCounter.failed != 1 {
			t.Errorf("counter value wasn't raised after the failed attempt to create a document. expected value: 1, actual value: %d",
				test.singleDocCreateCounter.failed)
		}
	} else {
		if test.singleDocCreateCounter.succeeded != 1 {
			t.Errorf("counter value wasn't raised after document creation. expected value: 1, actual value: %d",
				test.singleDocCreateCounter.succeeded)
		}
	}
}

func insertDocumentOKBehaviour(
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
	expected_url := fmt.Sprintf("/_api/document/%s", CollectionName)
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

func TestInsertDocumentOK(t *testing.T) {
	checkInsertDocument(t, false, insertDocumentOKBehaviour)
}

func insertDocumentRetryBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  nil,
	}

	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 404},
		Err:  nil,
	}

	req = next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//check request method
	if req.Method != "POST" {
		t.Errorf("Got wrong method %s instead of POST.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s", CollectionName)
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

func TestInsertDocumentRetry(t *testing.T) {
	checkInsertDocument(t, false, insertDocumentRetryBehaviour)
}

func TestInsertDocumentTimeout(t *testing.T) {
	checkInsertDocument(t, true, only503ReplyBehaviour)
}

func insertDocumentErrorBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {

	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	//return an error
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 500},
		Err:  fmt.Errorf("Error"),
	}
}

func TestInsertDocumentError(t *testing.T) {
	checkInsertDocument(t, true, insertDocumentErrorBehaviour)
}

// readExistingDocument tests
// a template for testing the readExistingDocument function
func checkReadExistingDoc(t *testing.T, expectError bool, skipExpectedValueCheck bool, behaviour util.Behaviour) {
	savedBackOffTime := BackOffTime
	BackOffTime = backOffTimeForTesting // to speed up tests
	defer func() { BackOffTime = savedBackOffTime }()
	test := NewMockTest(util.NewMockClient(t, behaviour))
	err := test.readExistingDocument(CollectionName, ExpectedDocument, skipExpectedValueCheck)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected result from insertDocument: must return an error")
	}
	if expectError {
		if test.readExistingCounter.failed != 1 {
			t.Errorf("counter value wasn't raised after the failed attempt to create a document. expected value: 1, actual value: %d",
				test.readExistingCounter.failed)
		}
	} else {
		if test.readExistingCounter.succeeded != 1 {
			t.Errorf("counter value wasn't raised after document creation. expected value: 1, actual value: %d",
				test.readExistingCounter.succeeded)
		}
	}
}

func readExistingDocOKBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}
	//send a successfull response
	reqResultDoc := req.Result.(interface{}).(*BigDocument)
	reqResultDoc.Seed = ExpectedDocument.Seed
	reqResultDoc.Key = ExpectedDocument.Key
	reqResultDoc.Rev = ExpectedDocument.Rev
	reqResultDoc.UpdateCounter = ExpectedDocument.UpdateCounter
	reqResultDoc.Value = ExpectedDocument.Value
	reqResultDoc.Name = ExpectedDocument.Name
	reqResultDoc.Odd = ExpectedDocument.Odd
	reqResultDoc.Payload = ExpectedDocument.Payload
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDoc(t *testing.T) {
	checkReadExistingDoc(t, false, false, readExistingDocOKBehaviour)
}

func readExistingDocIncorrectValueBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}
	//send incorrect value in response
	reqResultDoc := req.Result.(*BigDocument)
	reqResultDoc.Seed = ExpectedDocument.Seed
	reqResultDoc.Key = "INCORRECT_KEY"
	reqResultDoc.Rev = ExpectedDocument.Rev
	reqResultDoc.UpdateCounter = ExpectedDocument.UpdateCounter
	reqResultDoc.Value = ExpectedDocument.Value
	reqResultDoc.Name = ExpectedDocument.Name
	reqResultDoc.Odd = ExpectedDocument.Odd
	reqResultDoc.Payload = ExpectedDocument.Payload
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDocIncorrectValue(t *testing.T) {
	checkReadExistingDoc(t, true, false, readExistingDocIncorrectValueBehaviour)
}

func readExistingDocErrorBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "GET" {
		t.Errorf("Got wrong method %s instead of GET.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}
	//send error in response
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 503},
		Err:  fmt.Errorf("Error"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestReadExistingDocError(t *testing.T) {
	checkReadExistingDoc(t, true, false, readExistingDocErrorBehaviour)
}

func readExistingDocTimeoutBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	for {
		// Get a request:
		req := next(ctx, t, requests, true)
		if req == nil {
			return
		}
		//check request method
		if req.Method != "GET" {
			t.Errorf("Got wrong method %s instead of GET.", req.Method)
		}
		//check URL
		expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
		if req.UrlPath != expected_url {
			t.Errorf("Got wrong URL path %s instead of %s",
				req.UrlPath, expected_url)
		}
		//send 503 response
		responses <- &util.MockResponse{
			Resp: util.ArangoResponse{StatusCode: 503},
			Err:  nil,
		}
	}
}

func TestReadExistingDocTimeout(t *testing.T) {
	checkReadExistingDoc(t, true, false, readExistingDocTimeoutBehaviour)
}

// updateExistingDocument tests
// a template for testing the updateExistingDocument function
func checkUpdateDoc(t *testing.T, expectError bool, behaviour util.Behaviour) {
	savedBackOffTime := BackOffTime
	BackOffTime = backOffTimeForTesting // to speed up tests
	defer func() { BackOffTime = savedBackOffTime }()
	test := NewMockTest(util.NewMockClient(t, behaviour))
	newDoc, err := test.updateExistingDocument(CollectionName, ExpectedDocument.TestDocument)
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	}
	if err == nil && expectError {
		t.Errorf("unexpected result from insertDocument: must return an error")
	}
	if expectError {
		if test.updateExistingCounter.failed != 1 {
			t.Errorf("counter value wasn't raised after the failed attempt to create a document. expected value: 1, actual value: %d",
				test.updateExistingCounter.failed)
		}
	} else {
		if test.updateExistingCounter.succeeded != 1 {
			t.Errorf("counter value wasn't raised after document creation. expected value: 1, actual value: %d",
				test.updateExistingCounter.succeeded)
		}
	}
	if !expectError && !(newDoc.UpdateCounter == ChangedDocument.UpdateCounter && newDoc.Rev == ChangedDocument.Rev) {
		t.Error("incorrect document returned")
	}
}

func updateDocOKBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: ChangedDocument.Rev},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateDocOK(t *testing.T) {
	checkUpdateDoc(t, false, updateDocOKBehaviour)
}

func updateDocErrorBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}

	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{},
		Err:  fmt.Errorf("Error"),
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateDocError(t *testing.T) {
	checkUpdateDoc(t, true, updateDocErrorBehaviour)
}

func updateDocResp412Behaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 412},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateDocRest412(t *testing.T) {
	checkUpdateDoc(t, true, updateDocResp412Behaviour)
}

func updateDocResp409Behaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

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
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	reqResultDoc := req.Result.(*BigDocument)
	reqResultDoc.Seed = ChangedDocument.Seed
	reqResultDoc.Key = ChangedDocument.Key
	reqResultDoc.Rev = ChangedDocument.Rev
	reqResultDoc.UpdateCounter = ChangedDocument.UpdateCounter
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: ChangedDocument.Rev},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateDocRest409(t *testing.T) {
	checkUpdateDoc(t, false, updateDocResp409Behaviour)
}

func updateDocChangedIncorrectlyBehaviour(
	ctx context.Context, t *testing.T,
	requests chan *util.MockRequest, responses chan *util.MockResponse) {
	// Get a request:
	req := next(ctx, t, requests, true)
	if req == nil {
		return
	}
	//check request method
	if req.Method != "PATCH" {
		t.Errorf("Got wrong method %s instead of PATCH.", req.Method)
	}
	//check URL
	expected_url := fmt.Sprintf("/_api/document/%s/%s", CollectionName, ExpectedDocument.Key)
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//respond with 409, which means there is an update conflict
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 409},
		Err:  nil,
	}

	// Now the SUT must attempt to read the document to check if it was changed or not.
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
	if req.UrlPath != expected_url {
		t.Errorf("Got wrong URL path %s instead of %s",
			req.UrlPath, expected_url)
	}

	//respond as if the document was changed by another request
	reqResultDoc := req.Result.(*BigDocument)
	reqResultDoc.Seed = 999
	reqResultDoc.Key = "invalid_key"
	reqResultDoc.Rev = "invalid_rev"
	reqResultDoc.UpdateCounter = 100
	responses <- &util.MockResponse{
		Resp: util.ArangoResponse{StatusCode: 200, Rev: ChangedDocument.Rev},
		Err:  nil,
	}

	// No more requests coming:
	next(ctx, t, requests, false)
}

func TestUpdateDocChangedIncorrectly(t *testing.T) {
	checkUpdateDoc(t, true, updateDocChangedIncorrectlyBehaviour)
}

func TestUpdateDocTimeOut(t *testing.T) {
	savedReadTimeout := ReadTimeout
	ReadTimeout = readTimeoutForTesting // to speed up tests
	defer func() { ReadTimeout = savedReadTimeout }()
	checkUpdateDoc(t, true, only503ReplyBehaviour)
}
