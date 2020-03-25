// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/aws/ecs/ecs.go

// Package mocks is a generated GoMock package.
package mocks

import (
	ecs "github.com/aws/aws-sdk-go/service/ecs"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockecsClient is a mock of ecsClient interface
type MockecsClient struct {
	ctrl     *gomock.Controller
	recorder *MockecsClientMockRecorder
}

// MockecsClientMockRecorder is the mock recorder for MockecsClient
type MockecsClientMockRecorder struct {
	mock *MockecsClient
}

// NewMockecsClient creates a new mock instance
func NewMockecsClient(ctrl *gomock.Controller) *MockecsClient {
	mock := &MockecsClient{ctrl: ctrl}
	mock.recorder = &MockecsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockecsClient) EXPECT() *MockecsClientMockRecorder {
	return m.recorder
}

// DescribeTasks mocks base method
func (m *MockecsClient) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeTasks", input)
	ret0, _ := ret[0].(*ecs.DescribeTasksOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeTasks indicates an expected call of DescribeTasks
func (mr *MockecsClientMockRecorder) DescribeTasks(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeTasks", reflect.TypeOf((*MockecsClient)(nil).DescribeTasks), input)
}

// DescribeTaskDefinition mocks base method
func (m *MockecsClient) DescribeTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeTaskDefinition", input)
	ret0, _ := ret[0].(*ecs.DescribeTaskDefinitionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeTaskDefinition indicates an expected call of DescribeTaskDefinition
func (mr *MockecsClientMockRecorder) DescribeTaskDefinition(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeTaskDefinition", reflect.TypeOf((*MockecsClient)(nil).DescribeTaskDefinition), input)
}

// DescribeServices mocks base method
func (m *MockecsClient) DescribeServices(input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeServices", input)
	ret0, _ := ret[0].(*ecs.DescribeServicesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeServices indicates an expected call of DescribeServices
func (mr *MockecsClientMockRecorder) DescribeServices(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeServices", reflect.TypeOf((*MockecsClient)(nil).DescribeServices), input)
}

// ListTasks mocks base method
func (m *MockecsClient) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTasks", input)
	ret0, _ := ret[0].(*ecs.ListTasksOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTasks indicates an expected call of ListTasks
func (mr *MockecsClientMockRecorder) ListTasks(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTasks", reflect.TypeOf((*MockecsClient)(nil).ListTasks), input)
}
