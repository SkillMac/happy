package happy

import (
	"./hBaseComponent"
	"./hECS"
	"runtime"
	"time"
)

var launch *hBaseComponent.LauncherComponent

//新建一个服务节点
func NewServerNode() *hBaseComponent.LauncherComponent {
	//构造运行时
	runtime := hEcs.NewRuntime(hEcs.Config{ThreadPoolSize: runtime.NumCPU()})
	runtime.UpdateFrameByInterval(time.Millisecond * 100)

	//构造启动器
	launcher := &hBaseComponent.LauncherComponent{}
	runtime.Root().AddComponent(launcher)
	return launcher
}

//获取默认节点
func DefaultServer() *hBaseComponent.LauncherComponent {
	if launch != nil {
		return launch
	}
	launch = NewServerNode()
	return launch
}
