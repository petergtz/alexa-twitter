// Code generated by pegomock. DO NOT EDIT.
// Source: github.com/petergtz/alexa-twitter/verbalizer (interfaces: TwitterStatusesClient)

package verbalizer_test

import (
	twitter "github.com/dghubble/go-twitter/twitter"
	pegomock "github.com/petergtz/pegomock"
	http "net/http"
	"reflect"
	"time"
)

type MockTwitterStatusesClient struct {
	fail func(message string, callerSkip ...int)
}

func NewMockTwitterStatusesClient(options ...pegomock.Option) *MockTwitterStatusesClient {
	mock := &MockTwitterStatusesClient{}
	for _, option := range options {
		option.Apply(mock)
	}
	return mock
}

func (mock *MockTwitterStatusesClient) SetFailHandler(fh pegomock.FailHandler) { mock.fail = fh }
func (mock *MockTwitterStatusesClient) FailHandler() pegomock.FailHandler      { return mock.fail }

func (mock *MockTwitterStatusesClient) Show(_param0 int64, _param1 *twitter.StatusShowParams) (*twitter.Tweet, *http.Response, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockTwitterStatusesClient().")
	}
	params := []pegomock.Param{_param0, _param1}
	result := pegomock.GetGenericMockFrom(mock).Invoke("Show", params, []reflect.Type{reflect.TypeOf((**twitter.Tweet)(nil)).Elem(), reflect.TypeOf((**http.Response)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 *twitter.Tweet
	var ret1 *http.Response
	var ret2 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(*twitter.Tweet)
		}
		if result[1] != nil {
			ret1 = result[1].(*http.Response)
		}
		if result[2] != nil {
			ret2 = result[2].(error)
		}
	}
	return ret0, ret1, ret2
}

func (mock *MockTwitterStatusesClient) VerifyWasCalledOnce() *VerifierMockTwitterStatusesClient {
	return &VerifierMockTwitterStatusesClient{
		mock:                   mock,
		invocationCountMatcher: pegomock.Times(1),
	}
}

func (mock *MockTwitterStatusesClient) VerifyWasCalled(invocationCountMatcher pegomock.Matcher) *VerifierMockTwitterStatusesClient {
	return &VerifierMockTwitterStatusesClient{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
	}
}

func (mock *MockTwitterStatusesClient) VerifyWasCalledInOrder(invocationCountMatcher pegomock.Matcher, inOrderContext *pegomock.InOrderContext) *VerifierMockTwitterStatusesClient {
	return &VerifierMockTwitterStatusesClient{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		inOrderContext:         inOrderContext,
	}
}

func (mock *MockTwitterStatusesClient) VerifyWasCalledEventually(invocationCountMatcher pegomock.Matcher, timeout time.Duration) *VerifierMockTwitterStatusesClient {
	return &VerifierMockTwitterStatusesClient{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		timeout:                timeout,
	}
}

type VerifierMockTwitterStatusesClient struct {
	mock                   *MockTwitterStatusesClient
	invocationCountMatcher pegomock.Matcher
	inOrderContext         *pegomock.InOrderContext
	timeout                time.Duration
}

func (verifier *VerifierMockTwitterStatusesClient) Show(_param0 int64, _param1 *twitter.StatusShowParams) *MockTwitterStatusesClient_Show_OngoingVerification {
	params := []pegomock.Param{_param0, _param1}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Show", params, verifier.timeout)
	return &MockTwitterStatusesClient_Show_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type MockTwitterStatusesClient_Show_OngoingVerification struct {
	mock              *MockTwitterStatusesClient
	methodInvocations []pegomock.MethodInvocation
}

func (c *MockTwitterStatusesClient_Show_OngoingVerification) GetCapturedArguments() (int64, *twitter.StatusShowParams) {
	_param0, _param1 := c.GetAllCapturedArguments()
	return _param0[len(_param0)-1], _param1[len(_param1)-1]
}

func (c *MockTwitterStatusesClient_Show_OngoingVerification) GetAllCapturedArguments() (_param0 []int64, _param1 []*twitter.StatusShowParams) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]int64, len(c.methodInvocations))
		for u, param := range params[0] {
			_param0[u] = param.(int64)
		}
		_param1 = make([]*twitter.StatusShowParams, len(c.methodInvocations))
		for u, param := range params[1] {
			_param1[u] = param.(*twitter.StatusShowParams)
		}
	}
	return
}
