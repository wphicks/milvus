// Code generated by mockery v2.15.0. DO NOT EDIT.

package segments

import (
	commonpb "github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	mock "github.com/stretchr/testify/mock"

	querypb "github.com/milvus-io/milvus/internal/proto/querypb"
)

// MockSegmentManager is an autogenerated mock type for the SegmentManager type
type MockSegmentManager struct {
	mock.Mock
}

type MockSegmentManager_Expecter struct {
	mock *mock.Mock
}

func (_m *MockSegmentManager) EXPECT() *MockSegmentManager_Expecter {
	return &MockSegmentManager_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: segmentID
func (_m *MockSegmentManager) Get(segmentID int64) Segment {
	ret := _m.Called(segmentID)

	var r0 Segment
	if rf, ok := ret.Get(0).(func(int64) Segment); ok {
		r0 = rf(segmentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Segment)
		}
	}

	return r0
}

// MockSegmentManager_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockSegmentManager_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - segmentID int64
func (_e *MockSegmentManager_Expecter) Get(segmentID interface{}) *MockSegmentManager_Get_Call {
	return &MockSegmentManager_Get_Call{Call: _e.mock.On("Get", segmentID)}
}

func (_c *MockSegmentManager_Get_Call) Run(run func(segmentID int64)) *MockSegmentManager_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64))
	})
	return _c
}

func (_c *MockSegmentManager_Get_Call) Return(_a0 Segment) *MockSegmentManager_Get_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetBy provides a mock function with given fields: filters
func (_m *MockSegmentManager) GetBy(filters ...SegmentFilter) []Segment {
	_va := make([]interface{}, len(filters))
	for _i := range filters {
		_va[_i] = filters[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []Segment
	if rf, ok := ret.Get(0).(func(...SegmentFilter) []Segment); ok {
		r0 = rf(filters...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Segment)
		}
	}

	return r0
}

// MockSegmentManager_GetBy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBy'
type MockSegmentManager_GetBy_Call struct {
	*mock.Call
}

// GetBy is a helper method to define mock.On call
//   - filters ...SegmentFilter
func (_e *MockSegmentManager_Expecter) GetBy(filters ...interface{}) *MockSegmentManager_GetBy_Call {
	return &MockSegmentManager_GetBy_Call{Call: _e.mock.On("GetBy",
		append([]interface{}{}, filters...)...)}
}

func (_c *MockSegmentManager_GetBy_Call) Run(run func(filters ...SegmentFilter)) *MockSegmentManager_GetBy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]SegmentFilter, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(SegmentFilter)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *MockSegmentManager_GetBy_Call) Return(_a0 []Segment) *MockSegmentManager_GetBy_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetGrowing provides a mock function with given fields: segmentID
func (_m *MockSegmentManager) GetGrowing(segmentID int64) Segment {
	ret := _m.Called(segmentID)

	var r0 Segment
	if rf, ok := ret.Get(0).(func(int64) Segment); ok {
		r0 = rf(segmentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Segment)
		}
	}

	return r0
}

// MockSegmentManager_GetGrowing_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGrowing'
type MockSegmentManager_GetGrowing_Call struct {
	*mock.Call
}

// GetGrowing is a helper method to define mock.On call
//   - segmentID int64
func (_e *MockSegmentManager_Expecter) GetGrowing(segmentID interface{}) *MockSegmentManager_GetGrowing_Call {
	return &MockSegmentManager_GetGrowing_Call{Call: _e.mock.On("GetGrowing", segmentID)}
}

func (_c *MockSegmentManager_GetGrowing_Call) Run(run func(segmentID int64)) *MockSegmentManager_GetGrowing_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64))
	})
	return _c
}

func (_c *MockSegmentManager_GetGrowing_Call) Return(_a0 Segment) *MockSegmentManager_GetGrowing_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetSealed provides a mock function with given fields: segmentID
func (_m *MockSegmentManager) GetSealed(segmentID int64) Segment {
	ret := _m.Called(segmentID)

	var r0 Segment
	if rf, ok := ret.Get(0).(func(int64) Segment); ok {
		r0 = rf(segmentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Segment)
		}
	}

	return r0
}

// MockSegmentManager_GetSealed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSealed'
type MockSegmentManager_GetSealed_Call struct {
	*mock.Call
}

// GetSealed is a helper method to define mock.On call
//   - segmentID int64
func (_e *MockSegmentManager_Expecter) GetSealed(segmentID interface{}) *MockSegmentManager_GetSealed_Call {
	return &MockSegmentManager_GetSealed_Call{Call: _e.mock.On("GetSealed", segmentID)}
}

func (_c *MockSegmentManager_GetSealed_Call) Run(run func(segmentID int64)) *MockSegmentManager_GetSealed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64))
	})
	return _c
}

func (_c *MockSegmentManager_GetSealed_Call) Return(_a0 Segment) *MockSegmentManager_GetSealed_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetWithType provides a mock function with given fields: segmentID, typ
func (_m *MockSegmentManager) GetWithType(segmentID int64, typ commonpb.SegmentState) Segment {
	ret := _m.Called(segmentID, typ)

	var r0 Segment
	if rf, ok := ret.Get(0).(func(int64, commonpb.SegmentState) Segment); ok {
		r0 = rf(segmentID, typ)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Segment)
		}
	}

	return r0
}

// MockSegmentManager_GetWithType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetWithType'
type MockSegmentManager_GetWithType_Call struct {
	*mock.Call
}

// GetWithType is a helper method to define mock.On call
//   - segmentID int64
//   - typ commonpb.SegmentState
func (_e *MockSegmentManager_Expecter) GetWithType(segmentID interface{}, typ interface{}) *MockSegmentManager_GetWithType_Call {
	return &MockSegmentManager_GetWithType_Call{Call: _e.mock.On("GetWithType", segmentID, typ)}
}

func (_c *MockSegmentManager_GetWithType_Call) Run(run func(segmentID int64, typ commonpb.SegmentState)) *MockSegmentManager_GetWithType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64), args[1].(commonpb.SegmentState))
	})
	return _c
}

func (_c *MockSegmentManager_GetWithType_Call) Return(_a0 Segment) *MockSegmentManager_GetWithType_Call {
	_c.Call.Return(_a0)
	return _c
}

// Put provides a mock function with given fields: segmentType, segments
func (_m *MockSegmentManager) Put(segmentType commonpb.SegmentState, segments ...Segment) {
	_va := make([]interface{}, len(segments))
	for _i := range segments {
		_va[_i] = segments[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, segmentType)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// MockSegmentManager_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type MockSegmentManager_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - segmentType commonpb.SegmentState
//   - segments ...Segment
func (_e *MockSegmentManager_Expecter) Put(segmentType interface{}, segments ...interface{}) *MockSegmentManager_Put_Call {
	return &MockSegmentManager_Put_Call{Call: _e.mock.On("Put",
		append([]interface{}{segmentType}, segments...)...)}
}

func (_c *MockSegmentManager_Put_Call) Run(run func(segmentType commonpb.SegmentState, segments ...Segment)) *MockSegmentManager_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]Segment, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(Segment)
			}
		}
		run(args[0].(commonpb.SegmentState), variadicArgs...)
	})
	return _c
}

func (_c *MockSegmentManager_Put_Call) Return() *MockSegmentManager_Put_Call {
	_c.Call.Return()
	return _c
}

// Remove provides a mock function with given fields: segmentID, scope
func (_m *MockSegmentManager) Remove(segmentID int64, scope querypb.DataScope) {
	_m.Called(segmentID, scope)
}

// MockSegmentManager_Remove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Remove'
type MockSegmentManager_Remove_Call struct {
	*mock.Call
}

// Remove is a helper method to define mock.On call
//   - segmentID int64
//   - scope querypb.DataScope
func (_e *MockSegmentManager_Expecter) Remove(segmentID interface{}, scope interface{}) *MockSegmentManager_Remove_Call {
	return &MockSegmentManager_Remove_Call{Call: _e.mock.On("Remove", segmentID, scope)}
}

func (_c *MockSegmentManager_Remove_Call) Run(run func(segmentID int64, scope querypb.DataScope)) *MockSegmentManager_Remove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64), args[1].(querypb.DataScope))
	})
	return _c
}

func (_c *MockSegmentManager_Remove_Call) Return() *MockSegmentManager_Remove_Call {
	_c.Call.Return()
	return _c
}

// RemoveBy provides a mock function with given fields: filters
func (_m *MockSegmentManager) RemoveBy(filters ...SegmentFilter) {
	_va := make([]interface{}, len(filters))
	for _i := range filters {
		_va[_i] = filters[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// MockSegmentManager_RemoveBy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveBy'
type MockSegmentManager_RemoveBy_Call struct {
	*mock.Call
}

// RemoveBy is a helper method to define mock.On call
//   - filters ...SegmentFilter
func (_e *MockSegmentManager_Expecter) RemoveBy(filters ...interface{}) *MockSegmentManager_RemoveBy_Call {
	return &MockSegmentManager_RemoveBy_Call{Call: _e.mock.On("RemoveBy",
		append([]interface{}{}, filters...)...)}
}

func (_c *MockSegmentManager_RemoveBy_Call) Run(run func(filters ...SegmentFilter)) *MockSegmentManager_RemoveBy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]SegmentFilter, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(SegmentFilter)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *MockSegmentManager_RemoveBy_Call) Return() *MockSegmentManager_RemoveBy_Call {
	_c.Call.Return()
	return _c
}

type mockConstructorTestingTNewMockSegmentManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockSegmentManager creates a new instance of MockSegmentManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockSegmentManager(t mockConstructorTestingTNewMockSegmentManager) *MockSegmentManager {
	mock := &MockSegmentManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
