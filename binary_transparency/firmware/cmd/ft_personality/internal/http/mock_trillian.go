// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/google/trillian-examples/binary_transparency/firmware/cmd/ft_personality/internal/http (interfaces: Trillian)

package http

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/google/trillian/types"
)

// MockTrillian is a mock of Trillian interface.
type MockTrillian struct {
	ctrl     *gomock.Controller
	recorder *MockTrillianMockRecorder
}

// MockTrillianMockRecorder is the mock recorder for MockTrillian.
type MockTrillianMockRecorder struct {
	mock *MockTrillian
}

// NewMockTrillian creates a new mock instance.
func NewMockTrillian(ctrl *gomock.Controller) *MockTrillian {
	mock := &MockTrillian{ctrl: ctrl}
	mock.recorder = &MockTrillianMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTrillian) EXPECT() *MockTrillianMockRecorder {
	return m.recorder
}

// AddSignedStatement mocks base method.
func (m *MockTrillian) AddSignedStatement(arg0 context.Context, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddSignedStatement", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddSignedStatement indicates an expected call of AddSignedStatement.
func (mr *MockTrillianMockRecorder) AddSignedStatement(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSignedStatement", reflect.TypeOf((*MockTrillian)(nil).AddSignedStatement), arg0, arg1)
}

// ConsistencyProof mocks base method.
func (m *MockTrillian) ConsistencyProof(arg0 context.Context, arg1, arg2 uint64) ([][]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConsistencyProof", arg0, arg1, arg2)
	ret0, _ := ret[0].([][]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConsistencyProof indicates an expected call of ConsistencyProof.
func (mr *MockTrillianMockRecorder) ConsistencyProof(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConsistencyProof", reflect.TypeOf((*MockTrillian)(nil).ConsistencyProof), arg0, arg1, arg2)
}

// FirmwareManifestAtIndex mocks base method.
func (m *MockTrillian) FirmwareManifestAtIndex(arg0 context.Context, arg1, arg2 uint64) ([]byte, [][]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FirmwareManifestAtIndex", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([][]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// FirmwareManifestAtIndex indicates an expected call of FirmwareManifestAtIndex.
func (mr *MockTrillianMockRecorder) FirmwareManifestAtIndex(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FirmwareManifestAtIndex", reflect.TypeOf((*MockTrillian)(nil).FirmwareManifestAtIndex), arg0, arg1, arg2)
}

// InclusionProofByHash mocks base method.
func (m *MockTrillian) InclusionProofByHash(arg0 context.Context, arg1 []byte, arg2 uint64) (uint64, [][]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InclusionProofByHash", arg0, arg1, arg2)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].([][]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// InclusionProofByHash indicates an expected call of InclusionProofByHash.
func (mr *MockTrillianMockRecorder) InclusionProofByHash(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InclusionProofByHash", reflect.TypeOf((*MockTrillian)(nil).InclusionProofByHash), arg0, arg1, arg2)
}

// Root mocks base method.
func (m *MockTrillian) Root() *types.LogRootV1 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Root")
	ret0, _ := ret[0].(*types.LogRootV1)
	return ret0
}

// Root indicates an expected call of Root.
func (mr *MockTrillianMockRecorder) Root() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Root", reflect.TypeOf((*MockTrillian)(nil).Root))
}
