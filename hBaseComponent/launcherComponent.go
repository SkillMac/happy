package hBaseComponent

import (
	"custom/happy/hActor"
	"custom/happy/hCluster"
	"custom/happy/hCommon"
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"custom/happy/hRpc"
	"custom/happy/hTimer"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var ErrServerNotInit = errors.New("server is not initialize")

/* 服务端启动组件 */
type LauncherComponent struct {
	hEcs.ComponentBase
	componentGroup *hCluster.ComponentGroups
	Config         *hConfig.ConfigComponent
	Close          chan struct{}
	checkHandler   func()
}

func (this *LauncherComponent) IsUnique() int {
	return hEcs.UNIQUE_TYPE_GLOBAL
}

func (this *LauncherComponent) Initialize() error {
	//新建server
	this.Close = make(chan struct{})
	this.componentGroup = &hCluster.ComponentGroups{}

	//读取配置文件，初始化配置
	this.Root().AddComponent(&hConfig.ConfigComponent{})

	//缓存配置文件
	this.Config = hConfig.Config

	//设置runtime工作线程
	this.Runtime().SetMaxThread(hConfig.Config.CommonConfig.RuntimeMaxWorker)

	// 初始化邮件
	hCommon.NewAutoEmail(&hConfig.Config.CustomConfig.Email)

	//rpc设置
	rpc.CallTimeout = time.Millisecond * time.Duration(hConfig.Config.ClusterConfig.RpcCallTimeout)
	rpc.Timeout = time.Millisecond * time.Duration(hConfig.Config.ClusterConfig.RpcTimeout)
	rpc.HeartInterval = time.Millisecond * time.Duration(hConfig.Config.ClusterConfig.RpcHeartBeatInterval)
	rpc.DebugMode = hConfig.Config.CommonConfig.Debug

	//log设置
	switch hConfig.Config.CommonConfig.LogMode {
	case hLog.DAILY:
		hLog.SetRollingDaily(hConfig.Config.CommonConfig.LogPath, hConfig.Config.ClusterConfig.AppName+".log")
	case hLog.ROLLFILE:
		hLog.SetRollingFile(hConfig.Config.CommonConfig.LogPath, hConfig.Config.ClusterConfig.AppName+".log", hConfig.Config.CommonConfig.LogFileSizeMax*hLog.MB, hConfig.Config.CommonConfig.LogFileMax)
	}
	hLog.SetLevel(hConfig.Config.CommonConfig.LogLevel)
	return nil
}

func (this *LauncherComponent) Serve() {
	//添加NodeComponent组件，使对象成为分布式节点
	this.Root().AddComponent(&hCluster.NodeComponent{})

	//添加ActorProxy组件，组织节点间的通信
	this.Root().AddComponent(&hActor.ActorProxyComponent{})

	//添加数据库连接组件, 用于数据库的连接
	this.Root().AddComponent((&ModleComponent{}))

	//添加组件到待选组件列表，默认添加master,child组件
	this.AddComponentGroup("master", []hEcs.IComponent{&hCluster.MasterComponent{}})
	this.AddComponentGroup("child", []hEcs.IComponent{&hCluster.ChildComponent{}})
	if hConfig.Config.ClusterConfig.IsLocationMode && len(hConfig.Config.ClusterConfig.Role) > 0 && hConfig.Config.ClusterConfig.Role[0] != "single" {
		this.AddComponentGroup("location", []hEcs.IComponent{&hCluster.LocationComponent{}})
	}

	//处理single模式
	if len(hConfig.Config.ClusterConfig.Role) == 0 || hConfig.Config.ClusterConfig.Role[0] == "single" {
		hConfig.Config.ClusterConfig.Role = this.componentGroup.AllGroupsName()
	}

	//添加基础组件组,一般通过组建组的定义决定服务器节点的服务角色
	err := this.componentGroup.AttachGroupsTo(hConfig.Config.ClusterConfig.Role, this.Root())
	if err != nil {
		hLog.Fatal(err)
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	/* 清理测试代码,ide关闭信号无法命中断点 */
	//go func() {
	//	time.Sleep(time.Second*2)
	//	this.Close<- struct {}{}
	//}()

	//等待服务器关闭，并执行停机处理
	select {
	case <-c:
	case <-this.Close:
	}

	// 检查连接数量 大于 0 继续服务知道所有的玩家退出游戏
	if this.checkHandler != nil && !hConfig.Config.CommonConfig.Debug {
		hLog.Info("检查连接数量")
		this.checkHandler()
	}

	hLog.Info("====== Start to close this server, do some cleaning now ...... ======")
	//do something else
	err = this.Root().Destroy()
	if err != nil {
		hLog.Error(err)
	}
	<-hTimer.After(time.Second)
	hLog.Info("====== Server is closed ======")
}

//覆盖节点信息
func (this *LauncherComponent) OverrideNodeDefine(nodeConfName string) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	if s, ok := this.Config.ClusterConfig.NodeDefine[nodeConfName]; ok {
		this.Config.ClusterConfig.LocalAddress = s.LocalAddress
		this.Config.ClusterConfig.Role = s.Role
	} else {
		panic(errors.New(fmt.Sprintf("this config name [ %s ] not defined", nodeConfName)))
	}
}

//覆盖节点端口
func (this *LauncherComponent) OverrideNodePort(port string) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	ip := strings.Split(this.Config.ClusterConfig.LocalAddress, ":")[0]
	this.Config.ClusterConfig.LocalAddress = fmt.Sprintf("%s:%s", ip, port)
}

//覆盖节点角色
func (this *LauncherComponent) OverrideNodeRoles(roles []string) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	this.Config.ClusterConfig.Role = roles
}

//添加一个组件组到组建组列表，不会立即添加到对象
func (this *LauncherComponent) AddComponentGroup(groupName string, group []hEcs.IComponent) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	this.componentGroup.AddGroup(groupName, group)
}

//添加多个组件组到组建组列表，不会立即添加到对象
func (this *LauncherComponent) AddComponentGroups(groups map[string][]hEcs.IComponent) error {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	for groupName, group := range groups {
		this.componentGroup.AddGroup(groupName, group)
	}
	return nil
}

func (this *LauncherComponent) RegisterGateCheckCloseFunc(handler func()) {
	this.checkHandler = handler
}
