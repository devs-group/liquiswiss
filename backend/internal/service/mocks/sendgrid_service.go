// Code generated by MockGen. DO NOT EDIT.
// Source: liquiswiss/internal/service/sendgrid_service (interfaces: ISendgridService)
//
// Generated by this command:
//
//	mockgen -package=mocks -destination ../mocks/sendgrid_service.go liquiswiss/internal/service/sendgrid_service ISendgridService
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	mail "github.com/sendgrid/sendgrid-go/helpers/mail"
	gomock "go.uber.org/mock/gomock"
)

// MockISendgridService is a mock of ISendgridService interface.
type MockISendgridService struct {
	ctrl     *gomock.Controller
	recorder *MockISendgridServiceMockRecorder
}

// MockISendgridServiceMockRecorder is the mock recorder for MockISendgridService.
type MockISendgridServiceMockRecorder struct {
	mock *MockISendgridService
}

// NewMockISendgridService creates a new mock instance.
func NewMockISendgridService(ctrl *gomock.Controller) *MockISendgridService {
	mock := &MockISendgridService{ctrl: ctrl}
	mock.recorder = &MockISendgridServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockISendgridService) EXPECT() *MockISendgridServiceMockRecorder {
	return m.recorder
}

// SendMail mocks base method.
func (m *MockISendgridService) SendMail(arg0, arg1 *mail.Email, arg2 string, arg3 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMail", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMail indicates an expected call of SendMail.
func (mr *MockISendgridServiceMockRecorder) SendMail(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMail", reflect.TypeOf((*MockISendgridService)(nil).SendMail), arg0, arg1, arg2, arg3)
}

// SendPasswordResetMail mocks base method.
func (m *MockISendgridService) SendPasswordResetMail(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendPasswordResetMail", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendPasswordResetMail indicates an expected call of SendPasswordResetMail.
func (mr *MockISendgridServiceMockRecorder) SendPasswordResetMail(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPasswordResetMail", reflect.TypeOf((*MockISendgridService)(nil).SendPasswordResetMail), arg0, arg1)
}

// SendRegistrationMail mocks base method.
func (m *MockISendgridService) SendRegistrationMail(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendRegistrationMail", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendRegistrationMail indicates an expected call of SendRegistrationMail.
func (mr *MockISendgridServiceMockRecorder) SendRegistrationMail(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendRegistrationMail", reflect.TypeOf((*MockISendgridService)(nil).SendRegistrationMail), arg0, arg1)
}