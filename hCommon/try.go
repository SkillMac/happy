package hCommon

import (
	"custom/happy/hLog"
	"errors"
	"runtime/debug"
)

func Try(task func(), catch ...func(error)) {
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str = r.(error).Error()
			case string:
				str = r.(string)
			}
			err := errors.New(str + "\n" + string(debug.Stack()))
			if len(catch) > 0 {
				catch[0](err)
			} else {
				hLog.Error(err)
			}
		}
	})()
	task()
}
