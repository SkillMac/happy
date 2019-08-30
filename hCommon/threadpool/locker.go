package threadpool

import (
	"../../hLog"
	"errors"
	"runtime/debug"
	"sync"
)

// Locker binds a single mutex locked update
type Locker struct {
	lock   *sync.Mutex
	action func(data interface{})
}

// ParseMessage executes the action on the locker in mutex locked context
func (locker *Locker) Invoke() {
	locker.lock.Lock()
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str = r.(error).Error()
			case string:
				str = r.(string)
			}
			err := errors.New(str + string(debug.Stack()))
			hLog.Error(err)
		}
		locker.lock.Unlock()
	})()
	locker.action(nil)
}

// InvokeWith executes the action on the locker in mutex locked context with an argument
func (locker *Locker) InvokeWith(data interface{}) {
	locker.lock.Lock()
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str = r.(error).Error()
			case string:
				str = r.(string)
			}
			err := errors.New(str + string(debug.Stack()))
			hLog.Error(err)
		}
		locker.lock.Unlock()
	})()
	locker.action(data)
}
