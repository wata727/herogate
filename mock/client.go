// Code generated by MockGen. DO NOT EDIT.
// Source: iface/client.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	options "github.com/wata727/herogate/api/options"
	log "github.com/wata727/herogate/log"
	reflect "reflect"
)

// MockClientInterface is a mock of ClientInterface interface
type MockClientInterface struct {
	ctrl     *gomock.Controller
	recorder *MockClientInterfaceMockRecorder
}

// MockClientInterfaceMockRecorder is the mock recorder for MockClientInterface
type MockClientInterfaceMockRecorder struct {
	mock *MockClientInterface
}

// NewMockClientInterface creates a new mock instance
func NewMockClientInterface(ctrl *gomock.Controller) *MockClientInterface {
	mock := &MockClientInterface{ctrl: ctrl}
	mock.recorder = &MockClientInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClientInterface) EXPECT() *MockClientInterfaceMockRecorder {
	return m.recorder
}

// CreateApp mocks base method
func (m *MockClientInterface) CreateApp(appName string) (string, string) {
	ret := m.ctrl.Call(m, "CreateApp", appName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// CreateApp indicates an expected call of CreateApp
func (mr *MockClientInterfaceMockRecorder) CreateApp(appName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateApp", reflect.TypeOf((*MockClientInterface)(nil).CreateApp), appName)
}

// GetAppCreationProgress mocks base method
func (m *MockClientInterface) GetAppCreationProgress(appName string) int {
	ret := m.ctrl.Call(m, "GetAppCreationProgress", appName)
	ret0, _ := ret[0].(int)
	return ret0
}

// GetAppCreationProgress indicates an expected call of GetAppCreationProgress
func (mr *MockClientInterfaceMockRecorder) GetAppCreationProgress(appName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAppCreationProgress", reflect.TypeOf((*MockClientInterface)(nil).GetAppCreationProgress), appName)
}

// DescribeLogs mocks base method
func (m *MockClientInterface) DescribeLogs(appName string, options *options.DescribeLogs) []*log.Log {
	ret := m.ctrl.Call(m, "DescribeLogs", appName, options)
	ret0, _ := ret[0].([]*log.Log)
	return ret0
}

// DescribeLogs indicates an expected call of DescribeLogs
func (mr *MockClientInterfaceMockRecorder) DescribeLogs(appName, options interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeLogs", reflect.TypeOf((*MockClientInterface)(nil).DescribeLogs), appName, options)
}
