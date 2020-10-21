// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/term/selector/selector.go

// Package mocks is a generated GoMock package.
package mocks

import (
	config "github.com/aws/copilot-cli/internal/pkg/config"
	prompt "github.com/aws/copilot-cli/internal/pkg/term/prompt"
	workspace "github.com/aws/copilot-cli/internal/pkg/workspace"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockPrompter is a mock of Prompter interface
type MockPrompter struct {
	ctrl     *gomock.Controller
	recorder *MockPrompterMockRecorder
}

// MockPrompterMockRecorder is the mock recorder for MockPrompter
type MockPrompterMockRecorder struct {
	mock *MockPrompter
}

// NewMockPrompter creates a new mock instance
func NewMockPrompter(ctrl *gomock.Controller) *MockPrompter {
	mock := &MockPrompter{ctrl: ctrl}
	mock.recorder = &MockPrompterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPrompter) EXPECT() *MockPrompterMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockPrompter) Get(message, help string, validator prompt.ValidatorFunc, promptOpts ...prompt.Option) (string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{message, help, validator}
	for _, a := range promptOpts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockPrompterMockRecorder) Get(message, help, validator interface{}, promptOpts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{message, help, validator}, promptOpts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockPrompter)(nil).Get), varargs...)
}

// SelectOne mocks base method
func (m *MockPrompter) SelectOne(message, help string, options []string, promptOpts ...prompt.Option) (string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{message, help, options}
	for _, a := range promptOpts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SelectOne", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectOne indicates an expected call of SelectOne
func (mr *MockPrompterMockRecorder) SelectOne(message, help, options interface{}, promptOpts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{message, help, options}, promptOpts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectOne", reflect.TypeOf((*MockPrompter)(nil).SelectOne), varargs...)
}

// MultiSelect mocks base method
func (m *MockPrompter) MultiSelect(message, help string, options []string, promptOpts ...prompt.Option) ([]string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{message, help, options}
	for _, a := range promptOpts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "MultiSelect", varargs...)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MultiSelect indicates an expected call of MultiSelect
func (mr *MockPrompterMockRecorder) MultiSelect(message, help, options interface{}, promptOpts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{message, help, options}, promptOpts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MultiSelect", reflect.TypeOf((*MockPrompter)(nil).MultiSelect), varargs...)
}

// Confirm mocks base method
func (m *MockPrompter) Confirm(message, help string, promptOpts ...prompt.Option) (bool, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{message, help}
	for _, a := range promptOpts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Confirm", varargs...)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Confirm indicates an expected call of Confirm
func (mr *MockPrompterMockRecorder) Confirm(message, help interface{}, promptOpts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{message, help}, promptOpts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Confirm", reflect.TypeOf((*MockPrompter)(nil).Confirm), varargs...)
}

// MockAppEnvLister is a mock of AppEnvLister interface
type MockAppEnvLister struct {
	ctrl     *gomock.Controller
	recorder *MockAppEnvListerMockRecorder
}

// MockAppEnvListerMockRecorder is the mock recorder for MockAppEnvLister
type MockAppEnvListerMockRecorder struct {
	mock *MockAppEnvLister
}

// NewMockAppEnvLister creates a new mock instance
func NewMockAppEnvLister(ctrl *gomock.Controller) *MockAppEnvLister {
	mock := &MockAppEnvLister{ctrl: ctrl}
	mock.recorder = &MockAppEnvListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAppEnvLister) EXPECT() *MockAppEnvListerMockRecorder {
	return m.recorder
}

// ListEnvironments mocks base method
func (m *MockAppEnvLister) ListEnvironments(appName string) ([]*config.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironments", appName)
	ret0, _ := ret[0].([]*config.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironments indicates an expected call of ListEnvironments
func (mr *MockAppEnvListerMockRecorder) ListEnvironments(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironments", reflect.TypeOf((*MockAppEnvLister)(nil).ListEnvironments), appName)
}

// ListApplications mocks base method
func (m *MockAppEnvLister) ListApplications() ([]*config.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListApplications")
	ret0, _ := ret[0].([]*config.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListApplications indicates an expected call of ListApplications
func (mr *MockAppEnvListerMockRecorder) ListApplications() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListApplications", reflect.TypeOf((*MockAppEnvLister)(nil).ListApplications))
}

// MockConfigWlLister is a mock of ConfigWlLister interface
type MockConfigWlLister struct {
	ctrl     *gomock.Controller
	recorder *MockConfigWlListerMockRecorder
}

// MockConfigWlListerMockRecorder is the mock recorder for MockConfigWlLister
type MockConfigWlListerMockRecorder struct {
	mock *MockConfigWlLister
}

// NewMockConfigWlLister creates a new mock instance
func NewMockConfigWlLister(ctrl *gomock.Controller) *MockConfigWlLister {
	mock := &MockConfigWlLister{ctrl: ctrl}
	mock.recorder = &MockConfigWlListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigWlLister) EXPECT() *MockConfigWlListerMockRecorder {
	return m.recorder
}

// ListServices mocks base method
func (m *MockConfigWlLister) ListServices(appName string) ([]*config.Workload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServices", appName)
	ret0, _ := ret[0].([]*config.Workload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListServices indicates an expected call of ListServices
func (mr *MockConfigWlListerMockRecorder) ListServices(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServices", reflect.TypeOf((*MockConfigWlLister)(nil).ListServices), appName)
}

// ListJobs mocks base method
func (m *MockConfigWlLister) ListJobs(appName string) ([]*config.Workload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobs", appName)
	ret0, _ := ret[0].([]*config.Workload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobs indicates an expected call of ListJobs
func (mr *MockConfigWlListerMockRecorder) ListJobs(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobs", reflect.TypeOf((*MockConfigWlLister)(nil).ListJobs), appName)
}

// MockConfigLister is a mock of ConfigLister interface
type MockConfigLister struct {
	ctrl     *gomock.Controller
	recorder *MockConfigListerMockRecorder
}

// MockConfigListerMockRecorder is the mock recorder for MockConfigLister
type MockConfigListerMockRecorder struct {
	mock *MockConfigLister
}

// NewMockConfigLister creates a new mock instance
func NewMockConfigLister(ctrl *gomock.Controller) *MockConfigLister {
	mock := &MockConfigLister{ctrl: ctrl}
	mock.recorder = &MockConfigListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigLister) EXPECT() *MockConfigListerMockRecorder {
	return m.recorder
}

// ListEnvironments mocks base method
func (m *MockConfigLister) ListEnvironments(appName string) ([]*config.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironments", appName)
	ret0, _ := ret[0].([]*config.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironments indicates an expected call of ListEnvironments
func (mr *MockConfigListerMockRecorder) ListEnvironments(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironments", reflect.TypeOf((*MockConfigLister)(nil).ListEnvironments), appName)
}

// ListApplications mocks base method
func (m *MockConfigLister) ListApplications() ([]*config.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListApplications")
	ret0, _ := ret[0].([]*config.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListApplications indicates an expected call of ListApplications
func (mr *MockConfigListerMockRecorder) ListApplications() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListApplications", reflect.TypeOf((*MockConfigLister)(nil).ListApplications))
}

// ListServices mocks base method
func (m *MockConfigLister) ListServices(appName string) ([]*config.Workload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServices", appName)
	ret0, _ := ret[0].([]*config.Workload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListServices indicates an expected call of ListServices
func (mr *MockConfigListerMockRecorder) ListServices(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServices", reflect.TypeOf((*MockConfigLister)(nil).ListServices), appName)
}

// ListJobs mocks base method
func (m *MockConfigLister) ListJobs(appName string) ([]*config.Workload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobs", appName)
	ret0, _ := ret[0].([]*config.Workload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobs indicates an expected call of ListJobs
func (mr *MockConfigListerMockRecorder) ListJobs(appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobs", reflect.TypeOf((*MockConfigLister)(nil).ListJobs), appName)
}

// MockWsWorkloadLister is a mock of WsWorkloadLister interface
type MockWsWorkloadLister struct {
	ctrl     *gomock.Controller
	recorder *MockWsWorkloadListerMockRecorder
}

// MockWsWorkloadListerMockRecorder is the mock recorder for MockWsWorkloadLister
type MockWsWorkloadListerMockRecorder struct {
	mock *MockWsWorkloadLister
}

// NewMockWsWorkloadLister creates a new mock instance
func NewMockWsWorkloadLister(ctrl *gomock.Controller) *MockWsWorkloadLister {
	mock := &MockWsWorkloadLister{ctrl: ctrl}
	mock.recorder = &MockWsWorkloadListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWsWorkloadLister) EXPECT() *MockWsWorkloadListerMockRecorder {
	return m.recorder
}

// ServiceNames mocks base method
func (m *MockWsWorkloadLister) ServiceNames() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServiceNames")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServiceNames indicates an expected call of ServiceNames
func (mr *MockWsWorkloadListerMockRecorder) ServiceNames() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServiceNames", reflect.TypeOf((*MockWsWorkloadLister)(nil).ServiceNames))
}

// JobNames mocks base method
func (m *MockWsWorkloadLister) JobNames() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "JobNames")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// JobNames indicates an expected call of JobNames
func (mr *MockWsWorkloadListerMockRecorder) JobNames() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JobNames", reflect.TypeOf((*MockWsWorkloadLister)(nil).JobNames))
}

// MockWorkspaceRetriever is a mock of WorkspaceRetriever interface
type MockWorkspaceRetriever struct {
	ctrl     *gomock.Controller
	recorder *MockWorkspaceRetrieverMockRecorder
}

// MockWorkspaceRetrieverMockRecorder is the mock recorder for MockWorkspaceRetriever
type MockWorkspaceRetrieverMockRecorder struct {
	mock *MockWorkspaceRetriever
}

// NewMockWorkspaceRetriever creates a new mock instance
func NewMockWorkspaceRetriever(ctrl *gomock.Controller) *MockWorkspaceRetriever {
	mock := &MockWorkspaceRetriever{ctrl: ctrl}
	mock.recorder = &MockWorkspaceRetrieverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWorkspaceRetriever) EXPECT() *MockWorkspaceRetrieverMockRecorder {
	return m.recorder
}

// ServiceNames mocks base method
func (m *MockWorkspaceRetriever) ServiceNames() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServiceNames")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServiceNames indicates an expected call of ServiceNames
func (mr *MockWorkspaceRetrieverMockRecorder) ServiceNames() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServiceNames", reflect.TypeOf((*MockWorkspaceRetriever)(nil).ServiceNames))
}

// JobNames mocks base method
func (m *MockWorkspaceRetriever) JobNames() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "JobNames")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// JobNames indicates an expected call of JobNames
func (mr *MockWorkspaceRetrieverMockRecorder) JobNames() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JobNames", reflect.TypeOf((*MockWorkspaceRetriever)(nil).JobNames))
}

// Summary mocks base method
func (m *MockWorkspaceRetriever) Summary() (*workspace.Summary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Summary")
	ret0, _ := ret[0].(*workspace.Summary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Summary indicates an expected call of Summary
func (mr *MockWorkspaceRetrieverMockRecorder) Summary() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Summary", reflect.TypeOf((*MockWorkspaceRetriever)(nil).Summary))
}

// ListDockerfiles mocks base method
func (m *MockWorkspaceRetriever) ListDockerfiles() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDockerfiles")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDockerfiles indicates an expected call of ListDockerfiles
func (mr *MockWorkspaceRetrieverMockRecorder) ListDockerfiles() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDockerfiles", reflect.TypeOf((*MockWorkspaceRetriever)(nil).ListDockerfiles))
}

// MockDeployStoreClient is a mock of DeployStoreClient interface
type MockDeployStoreClient struct {
	ctrl     *gomock.Controller
	recorder *MockDeployStoreClientMockRecorder
}

// MockDeployStoreClientMockRecorder is the mock recorder for MockDeployStoreClient
type MockDeployStoreClientMockRecorder struct {
	mock *MockDeployStoreClient
}

// NewMockDeployStoreClient creates a new mock instance
func NewMockDeployStoreClient(ctrl *gomock.Controller) *MockDeployStoreClient {
	mock := &MockDeployStoreClient{ctrl: ctrl}
	mock.recorder = &MockDeployStoreClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeployStoreClient) EXPECT() *MockDeployStoreClientMockRecorder {
	return m.recorder
}

// ListDeployedServices mocks base method
func (m *MockDeployStoreClient) ListDeployedServices(appName, envName string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDeployedServices", appName, envName)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDeployedServices indicates an expected call of ListDeployedServices
func (mr *MockDeployStoreClientMockRecorder) ListDeployedServices(appName, envName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDeployedServices", reflect.TypeOf((*MockDeployStoreClient)(nil).ListDeployedServices), appName, envName)
}

// IsServiceDeployed mocks base method
func (m *MockDeployStoreClient) IsServiceDeployed(appName, envName, svcName string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsServiceDeployed", appName, envName, svcName)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsServiceDeployed indicates an expected call of IsServiceDeployed
func (mr *MockDeployStoreClientMockRecorder) IsServiceDeployed(appName, envName, svcName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsServiceDeployed", reflect.TypeOf((*MockDeployStoreClient)(nil).IsServiceDeployed), appName, envName, svcName)
}
