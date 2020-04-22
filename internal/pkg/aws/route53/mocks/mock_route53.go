// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/pkg/aws/route53/route53.go

// Package mocks is a generated GoMock package.
package mocks

import (
	route53 "github.com/aws/aws-sdk-go/service/route53"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// Mockapi is a mock of api interface
type Mockapi struct {
	ctrl     *gomock.Controller
	recorder *MockapiMockRecorder
}

// MockapiMockRecorder is the mock recorder for Mockapi
type MockapiMockRecorder struct {
	mock *Mockapi
}

// NewMockapi creates a new mock instance
func NewMockapi(ctrl *gomock.Controller) *Mockapi {
	mock := &Mockapi{ctrl: ctrl}
	mock.recorder = &MockapiMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Mockapi) EXPECT() *MockapiMockRecorder {
	return m.recorder
}

// ListHostedZonesByName mocks base method
func (m *Mockapi) ListHostedZonesByName(in *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListHostedZonesByName", in)
	ret0, _ := ret[0].(*route53.ListHostedZonesByNameOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListHostedZonesByName indicates an expected call of ListHostedZonesByName
func (mr *MockapiMockRecorder) ListHostedZonesByName(in interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListHostedZonesByName", reflect.TypeOf((*Mockapi)(nil).ListHostedZonesByName), in)
}
