// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/cli/cli.go

// Package mocks is a generated GoMock package.
package mocks

import (
	archer "github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	ecr "github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/ecr"
	describe "github.com/aws/amazon-ecs-cli-v2/internal/pkg/describe"
	command "github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/command"
	session "github.com/aws/aws-sdk-go/aws/session"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockactionCommand is a mock of actionCommand interface
type MockactionCommand struct {
	ctrl     *gomock.Controller
	recorder *MockactionCommandMockRecorder
}

// MockactionCommandMockRecorder is the mock recorder for MockactionCommand
type MockactionCommandMockRecorder struct {
	mock *MockactionCommand
}

// NewMockactionCommand creates a new mock instance
func NewMockactionCommand(ctrl *gomock.Controller) *MockactionCommand {
	mock := &MockactionCommand{ctrl: ctrl}
	mock.recorder = &MockactionCommandMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockactionCommand) EXPECT() *MockactionCommandMockRecorder {
	return m.recorder
}

// Ask mocks base method
func (m *MockactionCommand) Ask() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ask")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ask indicates an expected call of Ask
func (mr *MockactionCommandMockRecorder) Ask() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ask", reflect.TypeOf((*MockactionCommand)(nil).Ask))
}

// Validate mocks base method
func (m *MockactionCommand) Validate() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate")
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockactionCommandMockRecorder) Validate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockactionCommand)(nil).Validate))
}

// Execute mocks base method
func (m *MockactionCommand) Execute() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute")
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockactionCommandMockRecorder) Execute() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockactionCommand)(nil).Execute))
}

// RecommendedActions mocks base method
func (m *MockactionCommand) RecommendedActions() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecommendedActions")
	ret0, _ := ret[0].([]string)
	return ret0
}

// RecommendedActions indicates an expected call of RecommendedActions
func (mr *MockactionCommandMockRecorder) RecommendedActions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecommendedActions", reflect.TypeOf((*MockactionCommand)(nil).RecommendedActions))
}

// MockprojectService is a mock of projectService interface
type MockprojectService struct {
	ctrl     *gomock.Controller
	recorder *MockprojectServiceMockRecorder
}

// MockprojectServiceMockRecorder is the mock recorder for MockprojectService
type MockprojectServiceMockRecorder struct {
	mock *MockprojectService
}

// NewMockprojectService creates a new mock instance
func NewMockprojectService(ctrl *gomock.Controller) *MockprojectService {
	mock := &MockprojectService{ctrl: ctrl}
	mock.recorder = &MockprojectServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockprojectService) EXPECT() *MockprojectServiceMockRecorder {
	return m.recorder
}

// ListProjects mocks base method
func (m *MockprojectService) ListProjects() ([]*archer.Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProjects")
	ret0, _ := ret[0].([]*archer.Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProjects indicates an expected call of ListProjects
func (mr *MockprojectServiceMockRecorder) ListProjects() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProjects", reflect.TypeOf((*MockprojectService)(nil).ListProjects))
}

// GetProject mocks base method
func (m *MockprojectService) GetProject(projectName string) (*archer.Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProject", projectName)
	ret0, _ := ret[0].(*archer.Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProject indicates an expected call of GetProject
func (mr *MockprojectServiceMockRecorder) GetProject(projectName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProject", reflect.TypeOf((*MockprojectService)(nil).GetProject), projectName)
}

// CreateProject mocks base method
func (m *MockprojectService) CreateProject(project *archer.Project) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProject", project)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateProject indicates an expected call of CreateProject
func (mr *MockprojectServiceMockRecorder) CreateProject(project interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProject", reflect.TypeOf((*MockprojectService)(nil).CreateProject), project)
}

// ListEnvironments mocks base method
func (m *MockprojectService) ListEnvironments(projectName string) ([]*archer.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironments", projectName)
	ret0, _ := ret[0].([]*archer.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironments indicates an expected call of ListEnvironments
func (mr *MockprojectServiceMockRecorder) ListEnvironments(projectName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironments", reflect.TypeOf((*MockprojectService)(nil).ListEnvironments), projectName)
}

// GetEnvironment mocks base method
func (m *MockprojectService) GetEnvironment(projectName, environmentName string) (*archer.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEnvironment", projectName, environmentName)
	ret0, _ := ret[0].(*archer.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEnvironment indicates an expected call of GetEnvironment
func (mr *MockprojectServiceMockRecorder) GetEnvironment(projectName, environmentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEnvironment", reflect.TypeOf((*MockprojectService)(nil).GetEnvironment), projectName, environmentName)
}

// CreateEnvironment mocks base method
func (m *MockprojectService) CreateEnvironment(env *archer.Environment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEnvironment", env)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateEnvironment indicates an expected call of CreateEnvironment
func (mr *MockprojectServiceMockRecorder) CreateEnvironment(env interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEnvironment", reflect.TypeOf((*MockprojectService)(nil).CreateEnvironment), env)
}

// DeleteEnvironment mocks base method
func (m *MockprojectService) DeleteEnvironment(projectName, environmentName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEnvironment", projectName, environmentName)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEnvironment indicates an expected call of DeleteEnvironment
func (mr *MockprojectServiceMockRecorder) DeleteEnvironment(projectName, environmentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEnvironment", reflect.TypeOf((*MockprojectService)(nil).DeleteEnvironment), projectName, environmentName)
}

// ListApplications mocks base method
func (m *MockprojectService) ListApplications(projectName string) ([]*archer.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListApplications", projectName)
	ret0, _ := ret[0].([]*archer.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListApplications indicates an expected call of ListApplications
func (mr *MockprojectServiceMockRecorder) ListApplications(projectName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListApplications", reflect.TypeOf((*MockprojectService)(nil).ListApplications), projectName)
}

// GetApplication mocks base method
func (m *MockprojectService) GetApplication(projectName, applicationName string) (*archer.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplication", projectName, applicationName)
	ret0, _ := ret[0].(*archer.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplication indicates an expected call of GetApplication
func (mr *MockprojectServiceMockRecorder) GetApplication(projectName, applicationName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplication", reflect.TypeOf((*MockprojectService)(nil).GetApplication), projectName, applicationName)
}

// CreateApplication mocks base method
func (m *MockprojectService) CreateApplication(app *archer.Application) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateApplication", app)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateApplication indicates an expected call of CreateApplication
func (mr *MockprojectServiceMockRecorder) CreateApplication(app interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateApplication", reflect.TypeOf((*MockprojectService)(nil).CreateApplication), app)
}

// DeleteApplication mocks base method
func (m *MockprojectService) DeleteApplication(projectName, appName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteApplication", projectName, appName)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteApplication indicates an expected call of DeleteApplication
func (mr *MockprojectServiceMockRecorder) DeleteApplication(projectName, appName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApplication", reflect.TypeOf((*MockprojectService)(nil).DeleteApplication), projectName, appName)
}

// MockecrService is a mock of ecrService interface
type MockecrService struct {
	ctrl     *gomock.Controller
	recorder *MockecrServiceMockRecorder
}

// MockecrServiceMockRecorder is the mock recorder for MockecrService
type MockecrServiceMockRecorder struct {
	mock *MockecrService
}

// NewMockecrService creates a new mock instance
func NewMockecrService(ctrl *gomock.Controller) *MockecrService {
	mock := &MockecrService{ctrl: ctrl}
	mock.recorder = &MockecrServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockecrService) EXPECT() *MockecrServiceMockRecorder {
	return m.recorder
}

// GetRepository mocks base method
func (m *MockecrService) GetRepository(name string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRepository", name)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRepository indicates an expected call of GetRepository
func (mr *MockecrServiceMockRecorder) GetRepository(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRepository", reflect.TypeOf((*MockecrService)(nil).GetRepository), name)
}

// GetECRAuth mocks base method
func (m *MockecrService) GetECRAuth() (ecr.Auth, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetECRAuth")
	ret0, _ := ret[0].(ecr.Auth)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetECRAuth indicates an expected call of GetECRAuth
func (mr *MockecrServiceMockRecorder) GetECRAuth() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetECRAuth", reflect.TypeOf((*MockecrService)(nil).GetECRAuth))
}

// MockdockerService is a mock of dockerService interface
type MockdockerService struct {
	ctrl     *gomock.Controller
	recorder *MockdockerServiceMockRecorder
}

// MockdockerServiceMockRecorder is the mock recorder for MockdockerService
type MockdockerServiceMockRecorder struct {
	mock *MockdockerService
}

// NewMockdockerService creates a new mock instance
func NewMockdockerService(ctrl *gomock.Controller) *MockdockerService {
	mock := &MockdockerService{ctrl: ctrl}
	mock.recorder = &MockdockerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockdockerService) EXPECT() *MockdockerServiceMockRecorder {
	return m.recorder
}

// Build mocks base method
func (m *MockdockerService) Build(uri, tag, path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Build", uri, tag, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Build indicates an expected call of Build
func (mr *MockdockerServiceMockRecorder) Build(uri, tag, path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Build", reflect.TypeOf((*MockdockerService)(nil).Build), uri, tag, path)
}

// Login mocks base method
func (m *MockdockerService) Login(uri, username, password string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", uri, username, password)
	ret0, _ := ret[0].(error)
	return ret0
}

// Login indicates an expected call of Login
func (mr *MockdockerServiceMockRecorder) Login(uri, username, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockdockerService)(nil).Login), uri, username, password)
}

// Push mocks base method
func (m *MockdockerService) Push(uri, tag string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", uri, tag)
	ret0, _ := ret[0].(error)
	return ret0
}

// Push indicates an expected call of Push
func (mr *MockdockerServiceMockRecorder) Push(uri, tag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockdockerService)(nil).Push), uri, tag)
}

// Mockrunner is a mock of runner interface
type Mockrunner struct {
	ctrl     *gomock.Controller
	recorder *MockrunnerMockRecorder
}

// MockrunnerMockRecorder is the mock recorder for Mockrunner
type MockrunnerMockRecorder struct {
	mock *Mockrunner
}

// NewMockrunner creates a new mock instance
func NewMockrunner(ctrl *gomock.Controller) *Mockrunner {
	mock := &Mockrunner{ctrl: ctrl}
	mock.recorder = &MockrunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Mockrunner) EXPECT() *MockrunnerMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *Mockrunner) Run(name string, args []string, options ...command.Option) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{name, args}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Run", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run
func (mr *MockrunnerMockRecorder) Run(name, args interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{name, args}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*Mockrunner)(nil).Run), varargs...)
}

// MockdefaultSessionProvider is a mock of defaultSessionProvider interface
type MockdefaultSessionProvider struct {
	ctrl     *gomock.Controller
	recorder *MockdefaultSessionProviderMockRecorder
}

// MockdefaultSessionProviderMockRecorder is the mock recorder for MockdefaultSessionProvider
type MockdefaultSessionProviderMockRecorder struct {
	mock *MockdefaultSessionProvider
}

// NewMockdefaultSessionProvider creates a new mock instance
func NewMockdefaultSessionProvider(ctrl *gomock.Controller) *MockdefaultSessionProvider {
	mock := &MockdefaultSessionProvider{ctrl: ctrl}
	mock.recorder = &MockdefaultSessionProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockdefaultSessionProvider) EXPECT() *MockdefaultSessionProviderMockRecorder {
	return m.recorder
}

// Default mocks base method
func (m *MockdefaultSessionProvider) Default() (*session.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Default")
	ret0, _ := ret[0].(*session.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Default indicates an expected call of Default
func (mr *MockdefaultSessionProviderMockRecorder) Default() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Default", reflect.TypeOf((*MockdefaultSessionProvider)(nil).Default))
}

// MockregionalSessionProvider is a mock of regionalSessionProvider interface
type MockregionalSessionProvider struct {
	ctrl     *gomock.Controller
	recorder *MockregionalSessionProviderMockRecorder
}

// MockregionalSessionProviderMockRecorder is the mock recorder for MockregionalSessionProvider
type MockregionalSessionProviderMockRecorder struct {
	mock *MockregionalSessionProvider
}

// NewMockregionalSessionProvider creates a new mock instance
func NewMockregionalSessionProvider(ctrl *gomock.Controller) *MockregionalSessionProvider {
	mock := &MockregionalSessionProvider{ctrl: ctrl}
	mock.recorder = &MockregionalSessionProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockregionalSessionProvider) EXPECT() *MockregionalSessionProviderMockRecorder {
	return m.recorder
}

// DefaultWithRegion mocks base method
func (m *MockregionalSessionProvider) DefaultWithRegion(region string) (*session.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DefaultWithRegion", region)
	ret0, _ := ret[0].(*session.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DefaultWithRegion indicates an expected call of DefaultWithRegion
func (mr *MockregionalSessionProviderMockRecorder) DefaultWithRegion(region interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DefaultWithRegion", reflect.TypeOf((*MockregionalSessionProvider)(nil).DefaultWithRegion), region)
}

// MocksessionFromRoleProvider is a mock of sessionFromRoleProvider interface
type MocksessionFromRoleProvider struct {
	ctrl     *gomock.Controller
	recorder *MocksessionFromRoleProviderMockRecorder
}

// MocksessionFromRoleProviderMockRecorder is the mock recorder for MocksessionFromRoleProvider
type MocksessionFromRoleProviderMockRecorder struct {
	mock *MocksessionFromRoleProvider
}

// NewMocksessionFromRoleProvider creates a new mock instance
func NewMocksessionFromRoleProvider(ctrl *gomock.Controller) *MocksessionFromRoleProvider {
	mock := &MocksessionFromRoleProvider{ctrl: ctrl}
	mock.recorder = &MocksessionFromRoleProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MocksessionFromRoleProvider) EXPECT() *MocksessionFromRoleProviderMockRecorder {
	return m.recorder
}

// FromRole mocks base method
func (m *MocksessionFromRoleProvider) FromRole(roleARN, region string) (*session.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FromRole", roleARN, region)
	ret0, _ := ret[0].(*session.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FromRole indicates an expected call of FromRole
func (mr *MocksessionFromRoleProviderMockRecorder) FromRole(roleARN, region interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FromRole", reflect.TypeOf((*MocksessionFromRoleProvider)(nil).FromRole), roleARN, region)
}

// MocksessionProvider is a mock of sessionProvider interface
type MocksessionProvider struct {
	ctrl     *gomock.Controller
	recorder *MocksessionProviderMockRecorder
}

// MocksessionProviderMockRecorder is the mock recorder for MocksessionProvider
type MocksessionProviderMockRecorder struct {
	mock *MocksessionProvider
}

// NewMocksessionProvider creates a new mock instance
func NewMocksessionProvider(ctrl *gomock.Controller) *MocksessionProvider {
	mock := &MocksessionProvider{ctrl: ctrl}
	mock.recorder = &MocksessionProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MocksessionProvider) EXPECT() *MocksessionProviderMockRecorder {
	return m.recorder
}

// Default mocks base method
func (m *MocksessionProvider) Default() (*session.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Default")
	ret0, _ := ret[0].(*session.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Default indicates an expected call of Default
func (mr *MocksessionProviderMockRecorder) Default() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Default", reflect.TypeOf((*MocksessionProvider)(nil).Default))
}

// DefaultWithRegion mocks base method
func (m *MocksessionProvider) DefaultWithRegion(region string) (*session.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DefaultWithRegion", region)
	ret0, _ := ret[0].(*session.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DefaultWithRegion indicates an expected call of DefaultWithRegion
func (mr *MocksessionProviderMockRecorder) DefaultWithRegion(region interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DefaultWithRegion", reflect.TypeOf((*MocksessionProvider)(nil).DefaultWithRegion), region)
}

// FromRole mocks base method
func (m *MocksessionProvider) FromRole(roleARN, region string) (*session.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FromRole", roleARN, region)
	ret0, _ := ret[0].(*session.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FromRole indicates an expected call of FromRole
func (mr *MocksessionProviderMockRecorder) FromRole(roleARN, region interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FromRole", reflect.TypeOf((*MocksessionProvider)(nil).FromRole), roleARN, region)
}

// MockresourceIdentifier is a mock of resourceIdentifier interface
type MockresourceIdentifier struct {
	ctrl     *gomock.Controller
	recorder *MockresourceIdentifierMockRecorder
}

// MockresourceIdentifierMockRecorder is the mock recorder for MockresourceIdentifier
type MockresourceIdentifierMockRecorder struct {
	mock *MockresourceIdentifier
}

// NewMockresourceIdentifier creates a new mock instance
func NewMockresourceIdentifier(ctrl *gomock.Controller) *MockresourceIdentifier {
	mock := &MockresourceIdentifier{ctrl: ctrl}
	mock.recorder = &MockresourceIdentifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockresourceIdentifier) EXPECT() *MockresourceIdentifierMockRecorder {
	return m.recorder
}

// URI mocks base method
func (m *MockresourceIdentifier) URI(envName string) (*describe.WebAppURI, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "URI", envName)
	ret0, _ := ret[0].(*describe.WebAppURI)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// URI indicates an expected call of URI
func (mr *MockresourceIdentifierMockRecorder) URI(envName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "URI", reflect.TypeOf((*MockresourceIdentifier)(nil).URI), envName)
}

// MockstoreReader is a mock of storeReader interface
type MockstoreReader struct {
	ctrl     *gomock.Controller
	recorder *MockstoreReaderMockRecorder
}

// MockstoreReaderMockRecorder is the mock recorder for MockstoreReader
type MockstoreReaderMockRecorder struct {
	mock *MockstoreReader
}

// NewMockstoreReader creates a new mock instance
func NewMockstoreReader(ctrl *gomock.Controller) *MockstoreReader {
	mock := &MockstoreReader{ctrl: ctrl}
	mock.recorder = &MockstoreReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockstoreReader) EXPECT() *MockstoreReaderMockRecorder {
	return m.recorder
}

// ListProjects mocks base method
func (m *MockstoreReader) ListProjects() ([]*archer.Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProjects")
	ret0, _ := ret[0].([]*archer.Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProjects indicates an expected call of ListProjects
func (mr *MockstoreReaderMockRecorder) ListProjects() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProjects", reflect.TypeOf((*MockstoreReader)(nil).ListProjects))
}

// GetProject mocks base method
func (m *MockstoreReader) GetProject(projectName string) (*archer.Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProject", projectName)
	ret0, _ := ret[0].(*archer.Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProject indicates an expected call of GetProject
func (mr *MockstoreReaderMockRecorder) GetProject(projectName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProject", reflect.TypeOf((*MockstoreReader)(nil).GetProject), projectName)
}

// ListEnvironments mocks base method
func (m *MockstoreReader) ListEnvironments(projectName string) ([]*archer.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEnvironments", projectName)
	ret0, _ := ret[0].([]*archer.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEnvironments indicates an expected call of ListEnvironments
func (mr *MockstoreReaderMockRecorder) ListEnvironments(projectName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEnvironments", reflect.TypeOf((*MockstoreReader)(nil).ListEnvironments), projectName)
}

// GetEnvironment mocks base method
func (m *MockstoreReader) GetEnvironment(projectName, environmentName string) (*archer.Environment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEnvironment", projectName, environmentName)
	ret0, _ := ret[0].(*archer.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEnvironment indicates an expected call of GetEnvironment
func (mr *MockstoreReaderMockRecorder) GetEnvironment(projectName, environmentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEnvironment", reflect.TypeOf((*MockstoreReader)(nil).GetEnvironment), projectName, environmentName)
}

// ListApplications mocks base method
func (m *MockstoreReader) ListApplications(projectName string) ([]*archer.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListApplications", projectName)
	ret0, _ := ret[0].([]*archer.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListApplications indicates an expected call of ListApplications
func (mr *MockstoreReaderMockRecorder) ListApplications(projectName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListApplications", reflect.TypeOf((*MockstoreReader)(nil).ListApplications), projectName)
}

// GetApplication mocks base method
func (m *MockstoreReader) GetApplication(projectName, applicationName string) (*archer.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplication", projectName, applicationName)
	ret0, _ := ret[0].(*archer.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplication indicates an expected call of GetApplication
func (mr *MockstoreReaderMockRecorder) GetApplication(projectName, applicationName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplication", reflect.TypeOf((*MockstoreReader)(nil).GetApplication), projectName, applicationName)
}
