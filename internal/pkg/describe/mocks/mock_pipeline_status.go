// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/describe/pipeline_status.go

// Package mocks is a generated GoMock package.
package mocks

import (
	codepipeline "github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/codepipeline"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockpipelineStateGetter is a mock of pipelineStateGetter interface
type MockpipelineStateGetter struct {
	ctrl     *gomock.Controller
	recorder *MockpipelineStateGetterMockRecorder
}

// MockpipelineStateGetterMockRecorder is the mock recorder for MockpipelineStateGetter
type MockpipelineStateGetterMockRecorder struct {
	mock *MockpipelineStateGetter
}

// NewMockpipelineStateGetter creates a new mock instance
func NewMockpipelineStateGetter(ctrl *gomock.Controller) *MockpipelineStateGetter {
	mock := &MockpipelineStateGetter{ctrl: ctrl}
	mock.recorder = &MockpipelineStateGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockpipelineStateGetter) EXPECT() *MockpipelineStateGetterMockRecorder {
	return m.recorder
}

// GetPipelineState mocks base method
func (m *MockpipelineStateGetter) GetPipelineState(pipelineName string) (*codepipeline.PipelineState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPipelineState", pipelineName)
	ret0, _ := ret[0].(*codepipeline.PipelineState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPipelineState indicates an expected call of GetPipelineState
func (mr *MockpipelineStateGetterMockRecorder) GetPipelineState(pipelineName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPipelineState", reflect.TypeOf((*MockpipelineStateGetter)(nil).GetPipelineState), pipelineName)
}
