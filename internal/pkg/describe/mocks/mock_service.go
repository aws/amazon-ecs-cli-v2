// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/describe/service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	ecs "github.com/aws/copilot-cli/internal/pkg/aws/ecs"
	config "github.com/aws/copilot-cli/internal/pkg/config"
	gomock "github.com/golang/mock/gomock"
)

// MockecsClient is a mock of ecsClient interface.
type MockecsClient struct {
	ctrl     *gomock.Controller
	recorder *MockecsClientMockRecorder
}

// MockecsClientMockRecorder is the mock recorder for MockecsClient.
type MockecsClientMockRecorder struct {
	mock *MockecsClient
}

// NewMockecsClient creates a new mock instance.
func NewMockecsClient(ctrl *gomock.Controller) *MockecsClient {
	mock := &MockecsClient{ctrl: ctrl}
	mock.recorder = &MockecsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockecsClient) EXPECT() *MockecsClientMockRecorder {
	return m.recorder
}

// NetworkConfiguration mocks base method.
func (m *MockecsClient) NetworkConfiguration(cluster, serviceName string) (*ecs.NetworkConfiguration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkConfiguration", cluster, serviceName)
	ret0, _ := ret[0].(*ecs.NetworkConfiguration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NetworkConfiguration indicates an expected call of NetworkConfiguration.
func (mr *MockecsClientMockRecorder) NetworkConfiguration(cluster, serviceName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkConfiguration", reflect.TypeOf((*MockecsClient)(nil).NetworkConfiguration), cluster, serviceName)
}

// TaskDefinition mocks base method.
func (m *MockecsClient) TaskDefinition(taskDefName string) (*ecs.TaskDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TaskDefinition", taskDefName)
	ret0, _ := ret[0].(*ecs.TaskDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TaskDefinition indicates an expected call of TaskDefinition.
func (mr *MockecsClientMockRecorder) TaskDefinition(taskDefName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TaskDefinition", reflect.TypeOf((*MockecsClient)(nil).TaskDefinition), taskDefName)
}

// MockclusterDescriber is a mock of clusterDescriber interface.
type MockclusterDescriber struct {
	ctrl     *gomock.Controller
	recorder *MockclusterDescriberMockRecorder
}

// MockclusterDescriberMockRecorder is the mock recorder for MockclusterDescriber.
type MockclusterDescriberMockRecorder struct {
	mock *MockclusterDescriber
}

// NewMockclusterDescriber creates a new mock instance.
func NewMockclusterDescriber(ctrl *gomock.Controller) *MockclusterDescriber {
	mock := &MockclusterDescriber{ctrl: ctrl}
	mock.recorder = &MockclusterDescriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockclusterDescriber) EXPECT() *MockclusterDescriberMockRecorder {
	return m.recorder
}

// ClusterARN mocks base method.
func (m *MockclusterDescriber) ClusterARN(app, env string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClusterARN", app, env)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ClusterARN indicates an expected call of ClusterARN.
func (mr *MockclusterDescriberMockRecorder) ClusterARN(app, env interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClusterARN", reflect.TypeOf((*MockclusterDescriber)(nil).ClusterARN), app, env)
}

// MockConfigStoreSvc is a mock of ConfigStoreSvc interface.
type MockConfigStoreSvc struct {
	ctrl     *gomock.Controller
	recorder *MockConfigStoreSvcMockRecorder
}

// MockConfigStoreSvcMockRecorder is the mock recorder for MockConfigStoreSvc.
type MockConfigStoreSvcMockRecorder struct {
	mock *MockConfigStoreSvc
}

// NewMockConfigStoreSvc creates a new mock instance.
func NewMockConfigStoreSvc(ctrl *gomock.Controller) *MockConfigStoreSvc {
	mock := &MockConfigStoreSvc{ctrl: ctrl}
	mock.recorder = &MockConfigStoreSvcMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConfigStoreSvc) EXPECT() *MockConfigStoreSvcMockRecorder {
	return m.recorder
}

// GetEnvironment mocks base method.
func (m *MockConfigStoreSvc) GetEnvironment(appName, environmentName string) (*config.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEnvironment", appName, environmentName)
	ret0, _ := ret[0].(*config.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEnvironment indicates an expected call of GetEnvironment.
func (mr *MockConfigStoreSvcMockRecorder) GetEnvironment(appName, environmentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEnvironment", reflect.TypeOf((*MockConfigStoreSvc)(nil).GetEnvironment), appName, environmentName)
}

// ListEnvironments mocks base method.
func (m *MockConfigStoreSvc) ListEnvironments(appName string) ([]*config.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironments", appName)
	ret0, _ := ret[0].([]*config.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironments indicates an expected call of ListEnvironments.
func (mr *MockConfigStoreSvcMockRecorder) ListEnvironments(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironments", reflect.TypeOf((*MockConfigStoreSvc)(nil).ListEnvironments), appName)
}

// ListServices mocks base method.
func (m *MockConfigStoreSvc) ListServices(appName string) ([]*config.Workload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServices", appName)
	ret0, _ := ret[0].([]*config.Workload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListServices indicates an expected call of ListServices.
func (mr *MockConfigStoreSvcMockRecorder) ListServices(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServices", reflect.TypeOf((*MockConfigStoreSvc)(nil).ListServices), appName)
}

// MockDeployedEnvServicesLister is a mock of DeployedEnvServicesLister interface.
type MockDeployedEnvServicesLister struct {
	ctrl     *gomock.Controller
	recorder *MockDeployedEnvServicesListerMockRecorder
}

// MockDeployedEnvServicesListerMockRecorder is the mock recorder for MockDeployedEnvServicesLister.
type MockDeployedEnvServicesListerMockRecorder struct {
	mock *MockDeployedEnvServicesLister
}

// NewMockDeployedEnvServicesLister creates a new mock instance.
func NewMockDeployedEnvServicesLister(ctrl *gomock.Controller) *MockDeployedEnvServicesLister {
	mock := &MockDeployedEnvServicesLister{ctrl: ctrl}
	mock.recorder = &MockDeployedEnvServicesListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeployedEnvServicesLister) EXPECT() *MockDeployedEnvServicesListerMockRecorder {
	return m.recorder
}

// ListDeployedServices mocks base method.
func (m *MockDeployedEnvServicesLister) ListDeployedServices(appName, envName string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDeployedServices", appName, envName)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDeployedServices indicates an expected call of ListDeployedServices.
func (mr *MockDeployedEnvServicesListerMockRecorder) ListDeployedServices(appName, envName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDeployedServices", reflect.TypeOf((*MockDeployedEnvServicesLister)(nil).ListDeployedServices), appName, envName)
}

// ListEnvironmentsDeployedTo mocks base method.
func (m *MockDeployedEnvServicesLister) ListEnvironmentsDeployedTo(appName, svcName string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironmentsDeployedTo", appName, svcName)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironmentsDeployedTo indicates an expected call of ListEnvironmentsDeployedTo.
func (mr *MockDeployedEnvServicesListerMockRecorder) ListEnvironmentsDeployedTo(appName, svcName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironmentsDeployedTo", reflect.TypeOf((*MockDeployedEnvServicesLister)(nil).ListEnvironmentsDeployedTo), appName, svcName)
}
