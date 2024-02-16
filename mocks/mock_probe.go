// Code generated by MockGen. DO NOT EDIT.
// Source: globalping/probe/probe.go
//
// Generated by this command:
//
//	mockgen -source globalping/probe/probe.go -destination mocks/mock_probe.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	probe "github.com/jsdelivr/globalping-cli/globalping/probe"
	gomock "go.uber.org/mock/gomock"
)

// MockProbe is a mock of Probe interface.
type MockProbe struct {
	ctrl     *gomock.Controller
	recorder *MockProbeMockRecorder
}

// MockProbeMockRecorder is the mock recorder for MockProbe.
type MockProbeMockRecorder struct {
	mock *MockProbe
}

// NewMockProbe creates a new mock instance.
func NewMockProbe(ctrl *gomock.Controller) *MockProbe {
	mock := &MockProbe{ctrl: ctrl}
	mock.recorder = &MockProbeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProbe) EXPECT() *MockProbeMockRecorder {
	return m.recorder
}

// DetectContainerEngine mocks base method.
func (m *MockProbe) DetectContainerEngine() (probe.ContainerEngine, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetectContainerEngine")
	ret0, _ := ret[0].(probe.ContainerEngine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DetectContainerEngine indicates an expected call of DetectContainerEngine.
func (mr *MockProbeMockRecorder) DetectContainerEngine() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetectContainerEngine", reflect.TypeOf((*MockProbe)(nil).DetectContainerEngine))
}

// InspectContainer mocks base method.
func (m *MockProbe) InspectContainer(containerEngine probe.ContainerEngine) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InspectContainer", containerEngine)
	ret0, _ := ret[0].(error)
	return ret0
}

// InspectContainer indicates an expected call of InspectContainer.
func (mr *MockProbeMockRecorder) InspectContainer(containerEngine any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InspectContainer", reflect.TypeOf((*MockProbe)(nil).InspectContainer), containerEngine)
}

// RunContainer mocks base method.
func (m *MockProbe) RunContainer(containerEngine probe.ContainerEngine) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunContainer", containerEngine)
	ret0, _ := ret[0].(error)
	return ret0
}

// RunContainer indicates an expected call of RunContainer.
func (mr *MockProbeMockRecorder) RunContainer(containerEngine any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunContainer", reflect.TypeOf((*MockProbe)(nil).RunContainer), containerEngine)
}
