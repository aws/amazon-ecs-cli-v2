// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/repository/repository.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	exec "github.com/aws/copilot-cli/internal/pkg/exec"
	gomock "github.com/golang/mock/gomock"
)

// MockContainerLoginBuildPusher is a mock of ContainerLoginBuildPusher interface.
type MockContainerLoginBuildPusher struct {
	ctrl     *gomock.Controller
	recorder *MockContainerLoginBuildPusherMockRecorder
}

// MockContainerLoginBuildPusherMockRecorder is the mock recorder for MockContainerLoginBuildPusher.
type MockContainerLoginBuildPusherMockRecorder struct {
	mock *MockContainerLoginBuildPusher
}

// NewMockContainerLoginBuildPusher creates a new mock instance.
func NewMockContainerLoginBuildPusher(ctrl *gomock.Controller) *MockContainerLoginBuildPusher {
	mock := &MockContainerLoginBuildPusher{ctrl: ctrl}
	mock.recorder = &MockContainerLoginBuildPusherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContainerLoginBuildPusher) EXPECT() *MockContainerLoginBuildPusherMockRecorder {
	return m.recorder
}

// Build mocks base method.
func (m *MockContainerLoginBuildPusher) Build(args *exec.BuildArguments) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Build", args)
	ret0, _ := ret[0].(error)
	return ret0
}

// Build indicates an expected call of Build.
func (mr *MockContainerLoginBuildPusherMockRecorder) Build(args interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Build", reflect.TypeOf((*MockContainerLoginBuildPusher)(nil).Build), args)
}

// IsEcrCredentialHelperEnabled mocks base method
func (m *MockContainerLoginBuildPusher) IsEcrCredentialHelperEnabled(uri string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsEcrCredentialHelperEnabled", uri)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsEcrCredentialHelperEnabled indicates an expected call of IsEcrCredentialHelperEnabled
func (mr *MockContainerLoginBuildPusherMockRecorder) IsEcrCredentialHelperEnabled(uri string) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsEcrCredentialHelperEnabled", reflect.TypeOf((*MockContainerLoginBuildPusher)(nil).IsEcrCredentialHelperEnabled), uri)
}

// Login mocks base method.
func (m *MockContainerLoginBuildPusher) Login(uri, username, password string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", uri, username, password)
	ret0, _ := ret[0].(error)
	return ret0
}

// Login indicates an expected call of Login.
func (mr *MockContainerLoginBuildPusherMockRecorder) Login(uri, username, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockContainerLoginBuildPusher)(nil).Login), uri, username, password)
}

// Push mocks base method.
func (m *MockContainerLoginBuildPusher) Push(uri string, tags ...string) (string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{uri}
	for _, a := range tags {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Push", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Push indicates an expected call of Push.
func (mr *MockContainerLoginBuildPusherMockRecorder) Push(uri interface{}, tags ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{uri}, tags...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockContainerLoginBuildPusher)(nil).Push), varargs...)
}

// MockRegistry is a mock of Registry interface.
type MockRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockRegistryMockRecorder
}

// MockRegistryMockRecorder is the mock recorder for MockRegistry.
type MockRegistryMockRecorder struct {
	mock *MockRegistry
}

// NewMockRegistry creates a new mock instance.
func NewMockRegistry(ctrl *gomock.Controller) *MockRegistry {
	mock := &MockRegistry{ctrl: ctrl}
	mock.recorder = &MockRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegistry) EXPECT() *MockRegistryMockRecorder {
	return m.recorder
}

// Auth mocks base method.
func (m *MockRegistry) Auth() (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Auth")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Auth indicates an expected call of Auth.
func (mr *MockRegistryMockRecorder) Auth() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Auth", reflect.TypeOf((*MockRegistry)(nil).Auth))
}

// RepositoryURI mocks base method.
func (m *MockRegistry) RepositoryURI(name string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RepositoryURI", name)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RepositoryURI indicates an expected call of RepositoryURI.
func (mr *MockRegistryMockRecorder) RepositoryURI(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RepositoryURI", reflect.TypeOf((*MockRegistry)(nil).RepositoryURI), name)
}
