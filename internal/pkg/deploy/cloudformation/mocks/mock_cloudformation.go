// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/deploy/cloudformation/cloudformation.go

// Package mocks is a generated GoMock package.
package mocks

import (
	cloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	cloudformation0 "github.com/aws/copilot-cli/internal/pkg/aws/cloudformation"
	stackset "github.com/aws/copilot-cli/internal/pkg/aws/cloudformation/stackset"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockStackConfiguration is a mock of StackConfiguration interface
type MockStackConfiguration struct {
	ctrl     *gomock.Controller
	recorder *MockStackConfigurationMockRecorder
}

// MockStackConfigurationMockRecorder is the mock recorder for MockStackConfiguration
type MockStackConfigurationMockRecorder struct {
	mock *MockStackConfiguration
}

// NewMockStackConfiguration creates a new mock instance
func NewMockStackConfiguration(ctrl *gomock.Controller) *MockStackConfiguration {
	mock := &MockStackConfiguration{ctrl: ctrl}
	mock.recorder = &MockStackConfigurationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStackConfiguration) EXPECT() *MockStackConfigurationMockRecorder {
	return m.recorder
}

// StackName mocks base method
func (m *MockStackConfiguration) StackName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StackName")
	ret0, _ := ret[0].(string)
	return ret0
}

// StackName indicates an expected call of StackName
func (mr *MockStackConfigurationMockRecorder) StackName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StackName", reflect.TypeOf((*MockStackConfiguration)(nil).StackName))
}

// Template mocks base method
func (m *MockStackConfiguration) Template() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Template")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Template indicates an expected call of Template
func (mr *MockStackConfigurationMockRecorder) Template() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Template", reflect.TypeOf((*MockStackConfiguration)(nil).Template))
}

// Parameters mocks base method
func (m *MockStackConfiguration) Parameters() ([]*cloudformation.Parameter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Parameters")
	ret0, _ := ret[0].([]*cloudformation.Parameter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Parameters indicates an expected call of Parameters
func (mr *MockStackConfigurationMockRecorder) Parameters() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Parameters", reflect.TypeOf((*MockStackConfiguration)(nil).Parameters))
}

// Tags mocks base method
func (m *MockStackConfiguration) Tags() []*cloudformation.Tag {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tags")
	ret0, _ := ret[0].([]*cloudformation.Tag)
	return ret0
}

// Tags indicates an expected call of Tags
func (mr *MockStackConfigurationMockRecorder) Tags() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tags", reflect.TypeOf((*MockStackConfiguration)(nil).Tags))
}

// MockcfnClient is a mock of cfnClient interface
type MockcfnClient struct {
	ctrl     *gomock.Controller
	recorder *MockcfnClientMockRecorder
}

// MockcfnClientMockRecorder is the mock recorder for MockcfnClient
type MockcfnClientMockRecorder struct {
	mock *MockcfnClient
}

// NewMockcfnClient creates a new mock instance
func NewMockcfnClient(ctrl *gomock.Controller) *MockcfnClient {
	mock := &MockcfnClient{ctrl: ctrl}
	mock.recorder = &MockcfnClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockcfnClient) EXPECT() *MockcfnClientMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockcfnClient) Create(arg0 *cloudformation0.Stack) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockcfnClientMockRecorder) Create(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockcfnClient)(nil).Create), arg0)
}

// CreateAndWait mocks base method
func (m *MockcfnClient) CreateAndWait(arg0 *cloudformation0.Stack) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAndWait", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateAndWait indicates an expected call of CreateAndWait
func (mr *MockcfnClientMockRecorder) CreateAndWait(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAndWait", reflect.TypeOf((*MockcfnClient)(nil).CreateAndWait), arg0)
}

// WaitForCreate mocks base method
func (m *MockcfnClient) WaitForCreate(stackName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForCreate", stackName)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForCreate indicates an expected call of WaitForCreate
func (mr *MockcfnClientMockRecorder) WaitForCreate(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForCreate", reflect.TypeOf((*MockcfnClient)(nil).WaitForCreate), stackName)
}

// Update mocks base method
func (m *MockcfnClient) Update(arg0 *cloudformation0.Stack) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockcfnClientMockRecorder) Update(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockcfnClient)(nil).Update), arg0)
}

// UpdateAndWait mocks base method
func (m *MockcfnClient) UpdateAndWait(arg0 *cloudformation0.Stack) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAndWait", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAndWait indicates an expected call of UpdateAndWait
func (mr *MockcfnClientMockRecorder) UpdateAndWait(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAndWait", reflect.TypeOf((*MockcfnClient)(nil).UpdateAndWait), arg0)
}

// WaitForUpdate mocks base method
func (m *MockcfnClient) WaitForUpdate(stackName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForUpdate", stackName)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForUpdate indicates an expected call of WaitForUpdate
func (mr *MockcfnClientMockRecorder) WaitForUpdate(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForUpdate", reflect.TypeOf((*MockcfnClient)(nil).WaitForUpdate), stackName)
}

// Delete mocks base method
func (m *MockcfnClient) Delete(stackName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", stackName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockcfnClientMockRecorder) Delete(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockcfnClient)(nil).Delete), stackName)
}

// DeleteAndWait mocks base method
func (m *MockcfnClient) DeleteAndWait(stackName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAndWait", stackName)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAndWait indicates an expected call of DeleteAndWait
func (mr *MockcfnClientMockRecorder) DeleteAndWait(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAndWait", reflect.TypeOf((*MockcfnClient)(nil).DeleteAndWait), stackName)
}

// DeleteAndWaitWithRoleARN mocks base method
func (m *MockcfnClient) DeleteAndWaitWithRoleARN(stackName, roleARN string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAndWaitWithRoleARN", stackName, roleARN)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAndWaitWithRoleARN indicates an expected call of DeleteAndWaitWithRoleARN
func (mr *MockcfnClientMockRecorder) DeleteAndWaitWithRoleARN(stackName, roleARN interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAndWaitWithRoleARN", reflect.TypeOf((*MockcfnClient)(nil).DeleteAndWaitWithRoleARN), stackName, roleARN)
}

// Describe mocks base method
func (m *MockcfnClient) Describe(stackName string) (*cloudformation0.StackDescription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Describe", stackName)
	ret0, _ := ret[0].(*cloudformation0.StackDescription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Describe indicates an expected call of Describe
func (mr *MockcfnClientMockRecorder) Describe(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Describe", reflect.TypeOf((*MockcfnClient)(nil).Describe), stackName)
}

// TemplateBody mocks base method
func (m *MockcfnClient) TemplateBody(stackName string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TemplateBody", stackName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TemplateBody indicates an expected call of TemplateBody
func (mr *MockcfnClientMockRecorder) TemplateBody(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TemplateBody", reflect.TypeOf((*MockcfnClient)(nil).TemplateBody), stackName)
}

// Events mocks base method
func (m *MockcfnClient) Events(stackName string) ([]cloudformation0.StackEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Events", stackName)
	ret0, _ := ret[0].([]cloudformation0.StackEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Events indicates an expected call of Events
func (mr *MockcfnClientMockRecorder) Events(stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Events", reflect.TypeOf((*MockcfnClient)(nil).Events), stackName)
}

// MockstackSetClient is a mock of stackSetClient interface
type MockstackSetClient struct {
	ctrl     *gomock.Controller
	recorder *MockstackSetClientMockRecorder
}

// MockstackSetClientMockRecorder is the mock recorder for MockstackSetClient
type MockstackSetClientMockRecorder struct {
	mock *MockstackSetClient
}

// NewMockstackSetClient creates a new mock instance
func NewMockstackSetClient(ctrl *gomock.Controller) *MockstackSetClient {
	mock := &MockstackSetClient{ctrl: ctrl}
	mock.recorder = &MockstackSetClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockstackSetClient) EXPECT() *MockstackSetClientMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockstackSetClient) Create(name, template string, opts ...stackset.CreateOrUpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{name, template}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Create", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockstackSetClientMockRecorder) Create(name, template interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{name, template}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockstackSetClient)(nil).Create), varargs...)
}

// CreateInstancesAndWait mocks base method
func (m *MockstackSetClient) CreateInstancesAndWait(name string, accounts, regions []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateInstancesAndWait", name, accounts, regions)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateInstancesAndWait indicates an expected call of CreateInstancesAndWait
func (mr *MockstackSetClientMockRecorder) CreateInstancesAndWait(name, accounts, regions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateInstancesAndWait", reflect.TypeOf((*MockstackSetClient)(nil).CreateInstancesAndWait), name, accounts, regions)
}

// UpdateAndWait mocks base method
func (m *MockstackSetClient) UpdateAndWait(name, template string, opts ...stackset.CreateOrUpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{name, template}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateAndWait", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAndWait indicates an expected call of UpdateAndWait
func (mr *MockstackSetClientMockRecorder) UpdateAndWait(name, template interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{name, template}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAndWait", reflect.TypeOf((*MockstackSetClient)(nil).UpdateAndWait), varargs...)
}

// Describe mocks base method
func (m *MockstackSetClient) Describe(name string) (stackset.Description, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Describe", name)
	ret0, _ := ret[0].(stackset.Description)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Describe indicates an expected call of Describe
func (mr *MockstackSetClientMockRecorder) Describe(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Describe", reflect.TypeOf((*MockstackSetClient)(nil).Describe), name)
}

// InstanceSummaries mocks base method
func (m *MockstackSetClient) InstanceSummaries(name string, opts ...stackset.InstanceSummariesOption) ([]stackset.InstanceSummary, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{name}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "InstanceSummaries", varargs...)
	ret0, _ := ret[0].([]stackset.InstanceSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InstanceSummaries indicates an expected call of InstanceSummaries
func (mr *MockstackSetClientMockRecorder) InstanceSummaries(name interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{name}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstanceSummaries", reflect.TypeOf((*MockstackSetClient)(nil).InstanceSummaries), varargs...)
}

// Delete mocks base method
func (m *MockstackSetClient) Delete(name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockstackSetClientMockRecorder) Delete(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockstackSetClient)(nil).Delete), name)
}
