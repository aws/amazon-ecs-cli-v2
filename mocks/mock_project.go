// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/archer/project.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	archer "github.com/aws/PRIVATE-amazon-ecs-archer/internal/pkg/archer"
	gomock "github.com/golang/mock/gomock"
)

// MockProjectStore is a mock of ProjectStore interface
type MockProjectStore struct {
	ctrl     *gomock.Controller
	recorder *MockProjectStoreMockRecorder
}

// MockProjectStoreMockRecorder is the mock recorder for MockProjectStore
type MockProjectStoreMockRecorder struct {
	mock *MockProjectStore
}

// NewMockProjectStore creates a new mock instance
func NewMockProjectStore(ctrl *gomock.Controller) *MockProjectStore {
	mock := &MockProjectStore{ctrl: ctrl}
	mock.recorder = &MockProjectStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProjectStore) EXPECT() *MockProjectStoreMockRecorder {
	return m.recorder
}

// ListProjects mocks base method
func (m *MockProjectStore) ListProjects() ([]*archer.Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProjects")
	ret0, _ := ret[0].([]*archer.Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProjects indicates an expected call of ListProjects
func (mr *MockProjectStoreMockRecorder) ListProjects() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProjects", reflect.TypeOf((*MockProjectStore)(nil).ListProjects))
}

// CreateProject mocks base method
func (m *MockProjectStore) CreateProject(project *archer.Project) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProject", project)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateProject indicates an expected call of CreateProject
func (mr *MockProjectStoreMockRecorder) CreateProject(project interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProject", reflect.TypeOf((*MockProjectStore)(nil).CreateProject), project)
}

// MockProjectLister is a mock of ProjectLister interface
type MockProjectLister struct {
	ctrl     *gomock.Controller
	recorder *MockProjectListerMockRecorder
}

// MockProjectListerMockRecorder is the mock recorder for MockProjectLister
type MockProjectListerMockRecorder struct {
	mock *MockProjectLister
}

// NewMockProjectLister creates a new mock instance
func NewMockProjectLister(ctrl *gomock.Controller) *MockProjectLister {
	mock := &MockProjectLister{ctrl: ctrl}
	mock.recorder = &MockProjectListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProjectLister) EXPECT() *MockProjectListerMockRecorder {
	return m.recorder
}

// ListProjects mocks base method
func (m *MockProjectLister) ListProjects() ([]*archer.Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProjects")
	ret0, _ := ret[0].([]*archer.Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProjects indicates an expected call of ListProjects
func (mr *MockProjectListerMockRecorder) ListProjects() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProjects", reflect.TypeOf((*MockProjectLister)(nil).ListProjects))
}

// MockProjectCreator is a mock of ProjectCreator interface
type MockProjectCreator struct {
	ctrl     *gomock.Controller
	recorder *MockProjectCreatorMockRecorder
}

// MockProjectCreatorMockRecorder is the mock recorder for MockProjectCreator
type MockProjectCreatorMockRecorder struct {
	mock *MockProjectCreator
}

// NewMockProjectCreator creates a new mock instance
func NewMockProjectCreator(ctrl *gomock.Controller) *MockProjectCreator {
	mock := &MockProjectCreator{ctrl: ctrl}
	mock.recorder = &MockProjectCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProjectCreator) EXPECT() *MockProjectCreatorMockRecorder {
	return m.recorder
}

// CreateProject mocks base method
func (m *MockProjectCreator) CreateProject(project *archer.Project) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProject", project)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateProject indicates an expected call of CreateProject
func (mr *MockProjectCreatorMockRecorder) CreateProject(project interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProject", reflect.TypeOf((*MockProjectCreator)(nil).CreateProject), project)
}
