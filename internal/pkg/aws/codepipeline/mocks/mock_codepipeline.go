// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/aws/codepipeline/codepipeline.go

// Package mocks is a generated GoMock package.
package mocks

import (
	codepipeline "github.com/aws/aws-sdk-go/service/codepipeline"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockcodepipelineClient is a mock of codepipelineClient interface
type MockcodepipelineClient struct {
	ctrl     *gomock.Controller
	recorder *MockcodepipelineClientMockRecorder
}

// MockcodepipelineClientMockRecorder is the mock recorder for MockcodepipelineClient
type MockcodepipelineClientMockRecorder struct {
	mock *MockcodepipelineClient
}

// NewMockcodepipelineClient creates a new mock instance
func NewMockcodepipelineClient(ctrl *gomock.Controller) *MockcodepipelineClient {
	mock := &MockcodepipelineClient{ctrl: ctrl}
	mock.recorder = &MockcodepipelineClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockcodepipelineClient) EXPECT() *MockcodepipelineClientMockRecorder {
	return m.recorder
}

// GetPipeline mocks base method
func (m *MockcodepipelineClient) GetPipeline(arg0 *codepipeline.GetPipelineInput) (*codepipeline.GetPipelineOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPipeline", arg0)
	ret0, _ := ret[0].(*codepipeline.GetPipelineOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPipeline indicates an expected call of GetPipeline
func (mr *MockcodepipelineClientMockRecorder) GetPipeline(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPipeline", reflect.TypeOf((*MockcodepipelineClient)(nil).GetPipeline), arg0)
}
