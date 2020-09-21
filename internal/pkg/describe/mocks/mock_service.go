// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/describe/service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	cloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	ecs "github.com/aws/copilot-cli/internal/pkg/aws/ecs"
	config "github.com/aws/copilot-cli/internal/pkg/config"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockstackAndResourcesDescriber is a mock of stackAndResourcesDescriber interface
type MockstackAndResourcesDescriber struct {
	ctrl     *gomock.Controller
	recorder *MockstackAndResourcesDescriberMockRecorder
}

// MockstackAndResourcesDescriberMockRecorder is the mock recorder for MockstackAndResourcesDescriber
type MockstackAndResourcesDescriberMockRecorder struct {
	mock *MockstackAndResourcesDescriber
}

// NewMockstackAndResourcesDescriber creates a new mock instance
func NewMockstackAndResourcesDescriber(ctrl *gomock.Controller) *MockstackAndResourcesDescriber {
	mock := &MockstackAndResourcesDescriber{ctrl: ctrl}
	mock.recorder = &MockstackAndResourcesDescriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockstackAndResourcesDescriber) EXPECT() *MockstackAndResourcesDescriberMockRecorder {
	return m.recorder
}

// Stack mocks base method
func (m *MockstackAndResourcesDescriber) Stack(stackName string) (*cloudformation.Stack, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stack", stackName)
	ret0, _ := ret[0].(*cloudformation.Stack)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stack indicates an expected call of Stack
func (mr *MockstackAndResourcesDescriberMockRecorder) Stack(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stack", reflect.TypeOf((*MockstackAndResourcesDescriber)(nil).Stack), stackName)
}

// StackResources mocks base method
func (m *MockstackAndResourcesDescriber) StackResources(stackName string) ([]*cloudformation.StackResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StackResources", stackName)
	ret0, _ := ret[0].([]*cloudformation.StackResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StackResources indicates an expected call of StackResources
func (mr *MockstackAndResourcesDescriberMockRecorder) StackResources(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StackResources", reflect.TypeOf((*MockstackAndResourcesDescriber)(nil).StackResources), stackName)
}

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

// TaskDefinition mocks base method
func (m *MockecsClient) TaskDefinition(taskDefName string) (*ecs.TaskDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TaskDefinition", taskDefName)
	ret0, _ := ret[0].(*ecs.TaskDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TaskDefinition indicates an expected call of TaskDefinition
func (mr *MockecsClientMockRecorder) TaskDefinition(taskDefName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TaskDefinition", reflect.TypeOf((*MockecsClient)(nil).TaskDefinition), taskDefName)
}

// MockConfigStoreSvc is a mock of ConfigStoreSvc interface
type MockConfigStoreSvc struct {
	ctrl     *gomock.Controller
	recorder *MockConfigStoreSvcMockRecorder
}

// MockConfigStoreSvcMockRecorder is the mock recorder for MockConfigStoreSvc
type MockConfigStoreSvcMockRecorder struct {
	mock *MockConfigStoreSvc
}

// NewMockConfigStoreSvc creates a new mock instance
func NewMockConfigStoreSvc(ctrl *gomock.Controller) *MockConfigStoreSvc {
	mock := &MockConfigStoreSvc{ctrl: ctrl}
	mock.recorder = &MockConfigStoreSvcMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigStoreSvc) EXPECT() *MockConfigStoreSvcMockRecorder {
	return m.recorder
}

// GetEnvironment mocks base method
func (m *MockConfigStoreSvc) GetEnvironment(appName, environmentName string) (*config.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEnvironment", appName, environmentName)
	ret0, _ := ret[0].(*config.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEnvironment indicates an expected call of GetEnvironment
func (mr *MockConfigStoreSvcMockRecorder) GetEnvironment(appName, environmentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEnvironment", reflect.TypeOf((*MockConfigStoreSvc)(nil).GetEnvironment), appName, environmentName)
}

// ListEnvironments mocks base method
func (m *MockConfigStoreSvc) ListEnvironments(appName string) ([]*config.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironments", appName)
	ret0, _ := ret[0].([]*config.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironments indicates an expected call of ListEnvironments
func (mr *MockConfigStoreSvcMockRecorder) ListEnvironments(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironments", reflect.TypeOf((*MockConfigStoreSvc)(nil).ListEnvironments), appName)
}

// ListWorkloads mocks base method
func (m *MockConfigStoreSvc) ListWorkloads(appName string) ([]*config.Workload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListWorkloads", appName)
	ret0, _ := ret[0].([]*config.Workload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListWorkloads indicates an expected call of ListWorkloads
func (mr *MockConfigStoreSvcMockRecorder) ListWorkloads(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListWorkloads", reflect.TypeOf((*MockConfigStoreSvc)(nil).ListWorkloads), appName)
}

// MockDeployedEnvServicesLister is a mock of DeployedEnvServicesLister interface
type MockDeployedEnvServicesLister struct {
	ctrl     *gomock.Controller
	recorder *MockDeployedEnvServicesListerMockRecorder
}

// MockDeployedEnvServicesListerMockRecorder is the mock recorder for MockDeployedEnvServicesLister
type MockDeployedEnvServicesListerMockRecorder struct {
	mock *MockDeployedEnvServicesLister
}

// NewMockDeployedEnvServicesLister creates a new mock instance
func NewMockDeployedEnvServicesLister(ctrl *gomock.Controller) *MockDeployedEnvServicesLister {
	mock := &MockDeployedEnvServicesLister{ctrl: ctrl}
	mock.recorder = &MockDeployedEnvServicesListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeployedEnvServicesLister) EXPECT() *MockDeployedEnvServicesListerMockRecorder {
	return m.recorder
}

// ListEnvironmentsDeployedTo mocks base method
func (m *MockDeployedEnvServicesLister) ListEnvironmentsDeployedTo(appName, svcName string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironmentsDeployedTo", appName, svcName)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironmentsDeployedTo indicates an expected call of ListEnvironmentsDeployedTo
func (mr *MockDeployedEnvServicesListerMockRecorder) ListEnvironmentsDeployedTo(appName, svcName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironmentsDeployedTo", reflect.TypeOf((*MockDeployedEnvServicesLister)(nil).ListEnvironmentsDeployedTo), appName, svcName)
}

// ListDeployedServices mocks base method
func (m *MockDeployedEnvServicesLister) ListDeployedServices(appName, envName string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDeployedServices", appName, envName)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDeployedServices indicates an expected call of ListDeployedServices
func (mr *MockDeployedEnvServicesListerMockRecorder) ListDeployedServices(appName, envName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDeployedServices", reflect.TypeOf((*MockDeployedEnvServicesLister)(nil).ListDeployedServices), appName, envName)
}
