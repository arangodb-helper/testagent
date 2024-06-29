package util

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

type MockRequest struct {
	Method      string
	UrlPath     string
	Query       url.Values
	Header      map[string]string
	Input       interface{}
	ContentType string
	Result      interface{}
}

type MockResponse struct {
	Resp ArangoResponse
	Err  error
}

type Behaviour = func(context.Context, *testing.T, chan *MockRequest, chan *MockResponse)

type MockClient struct {
	test         *testing.T
	cancel       context.CancelFunc
	ctx          context.Context
	requests     chan *MockRequest
	responses    chan *MockResponse
	behaviour    Behaviour
	Wg           sync.WaitGroup
	databaseName string
}

type MockListener struct {
	// so far a black hole
}

func (ml MockListener) ReportFailure(f test.Failure) {
}

func NewMockClient(t *testing.T, behaviour Behaviour) *MockClient {
	mockClient := &MockClient{
		test:         t,
		requests:     make(chan *MockRequest),
		responses:    make(chan *MockResponse),
		behaviour:    behaviour,
		Wg:           sync.WaitGroup{},
		databaseName: "_system",
	}
	mockClient.ctx, mockClient.cancel = context.WithCancel(context.Background())
	mockClient.Wg.Add(1)
	go func() {
		mockClient.behaviour(mockClient.ctx, mockClient.test, mockClient.requests, mockClient.responses)
		mockClient.Wg.Done()
	}()
	return mockClient
}

func (mc *MockClient) SetCoordinator(coordinatorURL string) error {
	return nil
}

func (mc *MockClient) Get(urlPath string, query url.Values,
	header map[string]string, result interface{},
	successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	req := MockRequest{
		Method:  "GET",
		UrlPath: urlPath, Query: query, Header: header, Result: result,
	}
	mc.requests <- &req
	resp := <-mc.responses
	return []ArangoResponse{resp.Resp}, []error{resp.Err}
}

func (mc *MockClient) Delete(urlPath string, query url.Values,
	header map[string]string,
	successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	req := MockRequest{
		Method:  "DELETE",
		UrlPath: urlPath, Query: query, Header: header,
	}
	mc.requests <- &req
	resp := <-mc.responses
	return []ArangoResponse{resp.Resp}, []error{resp.Err}
}

func (mc *MockClient) Post(urlPath string, query url.Values, header map[string]string,
	input interface{}, contentType string, result interface{},
	successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	req := MockRequest{
		Method:  "POST",
		UrlPath: urlPath, Query: query, Header: header, Input: input,
		ContentType: contentType, Result: result,
	}
	mc.requests <- &req
	resp := <-mc.responses
	return []ArangoResponse{resp.Resp}, []error{resp.Err}
}

func (mc *MockClient) Patch(urlPath string, query url.Values, header map[string]string,
	input interface{}, contentType string, result interface{},
	successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	req := MockRequest{
		Method:  "PATCH",
		UrlPath: urlPath, Query: query, Header: header, Input: input,
		ContentType: contentType, Result: result,
	}
	mc.requests <- &req
	resp := <-mc.responses
	return []ArangoResponse{resp.Resp}, []error{resp.Err}
}

func (mc *MockClient) Put(urlPath string, query url.Values, header map[string]string,
	input interface{}, contentType string, result interface{},
	successStatusCodes, failureStatusCodes []int,
	operationTimeout time.Duration, retries int) ([]ArangoResponse, []error) {
	req := MockRequest{
		Method:  "PUT",
		UrlPath: urlPath, Query: query, Header: header, Input: input,
		ContentType: contentType, Result: result,
	}
	mc.requests <- &req
	resp := <-mc.responses
	return []ArangoResponse{resp.Resp}, []error{resp.Err}
}

func (mc *MockClient) Shutdown() {
	mc.cancel()
	mc.Wg.Wait()
}

func (c *MockClient) UseDatabase(databaseName string) {
	c.databaseName = databaseName
}
