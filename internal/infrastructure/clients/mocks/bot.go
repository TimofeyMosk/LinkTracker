// Code generated by mockery v2.46.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Bot is an autogenerated mock type for the Bot type
type Bot struct {
	mock.Mock
}

type Bot_Expecter struct {
	mock *mock.Mock
}

func (_m *Bot) EXPECT() *Bot_Expecter {
	return &Bot_Expecter{mock: &_m.Mock}
}

// HandleMessage provides a mock function with given fields: id, text
func (_m *Bot) HandleMessage(id int64, text string) string {
	ret := _m.Called(id, text)

	if len(ret) == 0 {
		panic("no return value specified for HandleMessage")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(int64, string) string); ok {
		r0 = rf(id, text)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Bot_HandleMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HandleMessage'
type Bot_HandleMessage_Call struct {
	*mock.Call
}

// HandleMessage is a helper method to define mock.On call
//   - id int64
//   - text string
func (_e *Bot_Expecter) HandleMessage(id interface{}, text interface{}) *Bot_HandleMessage_Call {
	return &Bot_HandleMessage_Call{Call: _e.mock.On("HandleMessage", id, text)}
}

func (_c *Bot_HandleMessage_Call) Run(run func(id int64, text string)) *Bot_HandleMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64), args[1].(string))
	})
	return _c
}

func (_c *Bot_HandleMessage_Call) Return(_a0 string) *Bot_HandleMessage_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Bot_HandleMessage_Call) RunAndReturn(run func(int64, string) string) *Bot_HandleMessage_Call {
	_c.Call.Return(run)
	return _c
}

// NewBot creates a new instance of Bot. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBot(t interface {
	mock.TestingT
	Cleanup(func())
}) *Bot {
	mock := &Bot{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
