// Code generated by mockery v2.1.0. DO NOT EDIT.

package mocks

import (
	abcicli "github.com/HPISTechnologies/consensus-engine/abci/client"
	mock "github.com/stretchr/testify/mock"

	types "github.com/HPISTechnologies/consensus-engine/abci/types"
)

// AppConnMempool is an autogenerated mock type for the AppConnMempool type
type AppConnMempool struct {
	mock.Mock
}

// CheckTxAsync provides a mock function with given fields: _a0
func (_m *AppConnMempool) CheckTxAsync(_a0 types.RequestCheckTx) *abcicli.ReqRes {
	ret := _m.Called(_a0)

	var r0 *abcicli.ReqRes
	if rf, ok := ret.Get(0).(func(types.RequestCheckTx) *abcicli.ReqRes); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abcicli.ReqRes)
		}
	}

	return r0
}

// CheckTxSync provides a mock function with given fields: _a0
func (_m *AppConnMempool) CheckTxSync(_a0 types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	ret := _m.Called(_a0)

	var r0 *types.ResponseCheckTx
	if rf, ok := ret.Get(0).(func(types.RequestCheckTx) *types.ResponseCheckTx); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseCheckTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.RequestCheckTx) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Error provides a mock function with given fields:
func (_m *AppConnMempool) Error() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FlushAsync provides a mock function with given fields:
func (_m *AppConnMempool) FlushAsync() *abcicli.ReqRes {
	ret := _m.Called()

	var r0 *abcicli.ReqRes
	if rf, ok := ret.Get(0).(func() *abcicli.ReqRes); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abcicli.ReqRes)
		}
	}

	return r0
}

// FlushSync provides a mock function with given fields:
func (_m *AppConnMempool) FlushSync() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetResponseCallback provides a mock function with given fields: _a0
func (_m *AppConnMempool) SetResponseCallback(_a0 abcicli.Callback) {
	_m.Called(_a0)
}
