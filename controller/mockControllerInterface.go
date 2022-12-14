package controller

import (
	dtos "github.com/iamcathal/booksbooksbooks/dtos"
	html "golang.org/x/net/html"

	mock "github.com/stretchr/testify/mock"

	time "time"

	websocket "github.com/gorilla/websocket"
)

// MockCntrInterface is an autogenerated mock type for the MockCntrInterface type
type MockCntrInterface struct {
	mock.Mock
}

// DeliverWebhook provides a mock function with given fields: msg
func (_m *MockCntrInterface) DeliverWebhook(msg dtos.DiscordMsg) error {
	ret := _m.Called(msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(dtos.DiscordMsg) error); ok {
		r0 = rf(msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFormattedTime provides a mock function with given fields:
func (_m *MockCntrInterface) GetFormattedTime() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPage provides a mock function with given fields: url
func (_m *MockCntrInterface) GetPage(pageURL string) *html.Node {
	ret := _m.Called(pageURL)

	var r0 *html.Node
	if rf, ok := ret.Get(0).(func(string) *html.Node); ok {
		r0 = rf(pageURL)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*html.Node)
		}
	}

	return r0
}

// GetPage provides a mock function with given fields: url
func (_m *MockCntrInterface) Get(pageURL string) []byte {
	ret := _m.Called(pageURL)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(pageURL)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// Sleep provides a mock function with given fields: duration
func (_m *MockCntrInterface) Sleep(duration time.Duration) {
	_m.Called(duration)
}

// WriteWsMessage provides a mock function with given fields: msg, ws
func (_m *MockCntrInterface) WriteWsMessage(msg []byte, ws *websocket.Conn) error {
	ret := _m.Called(msg, ws)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, *websocket.Conn) error); ok {
		r0 = rf(msg, ws)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockCntrInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockCntrInterface creates a new instance of MockCntrInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockCntrInterface(t mockConstructorTestingTNewMockCntrInterface) *MockCntrInterface {
	mock := &MockCntrInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
