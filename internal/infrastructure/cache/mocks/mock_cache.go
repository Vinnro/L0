package mocks

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

type MockICache struct {
	ctrl     *gomock.Controller
	recorder *MockICacheMockRecorder
}
type MockICacheMockRecorder struct {
	mock *MockICache
}

func NewMockICache(ctrl *gomock.Controller) *MockICache {
	mock := &MockICache{ctrl: ctrl}
	mock.recorder = &MockICacheMockRecorder{mock}
	return mock
}
func (m *MockICache) EXPECT() *MockICacheMockRecorder {
	return m.recorder
}

// Del mocks base method.
func (m *MockICache) Del(key string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del", key)
	ret0, _ := ret[0].(bool)
	return ret0
}
func (mr *MockICacheMockRecorder) Del(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockICache)(nil).Del), key)
}
func (m *MockICache) Get(key string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", key)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockICacheMockRecorder) Get(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockICache)(nil).Get), key)
}

func (m *MockICache) Set(key string, value []byte, expire time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", key, value, expire)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockICacheMockRecorder) Set(key, value, expire interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockICache)(nil).Set), key, value, expire)
}
