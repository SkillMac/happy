package happy

import (
	"custom/happy/hBaseComponent"
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"fmt"
	"runtime"
	"time"
)

var launch *hBaseComponent.LauncherComponent

//新建一个服务节点
func NewServerNode(nodeConfigName string) *hBaseComponent.LauncherComponent {
	//构造运行时
	runtime := hEcs.NewRuntime(hEcs.Config{ThreadPoolSize: runtime.NumCPU()})
	runtime.UpdateFrameByInterval(time.Millisecond * 100)

	//构造启动器
	launcher := &hBaseComponent.LauncherComponent{}

	// 提前构造配置文件
	runtime.Root().AddComponent(&hConfig.ConfigComponent{})

	if nodeConfigName != "" {
		launcher.Config = hConfig.Config
		launcher.OverrideNodeDefine(nodeConfigName)
		hLog.Info(fmt.Sprintf("Override node info: [ %s ]", nodeConfigName))
	}

	runtime.Root().AddComponent(launcher)
	return launcher
}

//获取默认节点
func DefaultServer(nodeConfigName string) *hBaseComponent.LauncherComponent {
	if launch != nil {
		return launch
	}
	launch = NewServerNode(nodeConfigName)
	return launch
}
