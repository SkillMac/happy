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
	"sync"
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
	//checkHandler   func()
	waitGroup     *sync.WaitGroup
	nodeComponent *hCluster.NodeComponent
	master        *rpc.TcpClient
}

func (this *LauncherComponent) IsUnique() int {
	return hEcs.UNIQUE_TYPE_GLOBAL
}

func (this *LauncherComponent) Initialize() error {
	this.waitGroup = &sync.WaitGroup{}
	//新建server
	this.Close = make(chan struct{})
	this.componentGroup = &hCluster.ComponentGroups{}

	////读取配置文件，初始化配置
	//this.Root().AddComponent(&hConfig.ConfigComponent{})

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

	hLog.InitLogger(hConfig.Config.ClusterConfig.WorkName, hConfig.Config.CommonConfig.LogFileSizeMax, hConfig.Config.CommonConfig.LogFileMax,
		hConfig.Config.CommonConfig.LogLevel, hConfig.Config.CommonConfig.LogConsolePrint)

	return nil
}

func (this *LauncherComponent) registerService() {

	s := new(LauncherService)
	s.init(this)
	err := this.nodeComponent.Register(s)
	if err != nil {
		panic(err)
	}
}

func (this *LauncherComponent) Serve() {
	//添加NodeComponent组件，使对象成为分布式节点
	this.nodeComponent = &hCluster.NodeComponent{}
	this.Root().AddComponent(this.nodeComponent)

	this.CheckNodeStatus()
	this.registerService()

	//添加ActorProxy组件，组织节点间的通信
	this.Root().AddComponent(&hActor.ActorProxyComponent{})

	//添加组件到待选组件列表，默认添加master,child组件
	this.AddComponentGroup("master", []hEcs.IComponent{&hCluster.MasterComponent{}})
	childComponent := &hCluster.ChildComponent{}
	this.AddComponentGroup("child", []hEcs.IComponent{childComponent})
	if hConfig.Config.ClusterConfig.IsLocationMode && len(hConfig.Config.ClusterConfig.Role) > 0 && hConfig.Config.ClusterConfig.Role[0] != "single" {
		this.AddComponentGroup("location", []hEcs.IComponent{&hCluster.LocationComponent{}})
	}

	//处理single模式
	if len(hConfig.Config.ClusterConfig.Role) == 0 || hConfig.Config.ClusterConfig.Role[0] == "single" {
		hConfig.Config.ClusterConfig.Role = this.componentGroup.AllGroupsName()
		hConfig.Config.ClusterConfig.Role = append(hConfig.Config.ClusterConfig.Role, "modle")
	}

	// 添加数据库连接
	if hCommon.Contains(this.Config.ClusterConfig.Role, "modle") {
		// 有 gate 角色才添加
		this.AddComponentGroup("modle", []hEcs.IComponent{&ModleComponent{}})
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
	if !hConfig.Config.CommonConfig.Debug {
		if childComponent != nil {
			this.Root().RemoveComponent(childComponent)
		}
		//if this.checkHandler != nil {
		//	hLog.Info("检查连接数量")
		//	this.checkHandler()
		//}

		// 检查所有组件服务的状态
		hLog.Info("检查Root节点下的服务组件是够服务完毕")
		for i := 0; i < len(this.Config.ClusterConfig.Role); i++ {
			obj, err := this.Root().GetObject(this.Config.ClusterConfig.Role[i])
			if err != nil {
				continue
			}

			allCompts := obj.AllComponents()
			for component, err := allCompts.Next(); err == nil; component, err = allCompts.Next() {
				hCommon.Try(func() {
					if comptL, ok := component.(IBDestroy); ok {
						comptL.CheckClose(this.waitGroup)
					}
				})
			}
		}

		hLog.Info("检查所有组件是否服务完毕")
		allComponents := this.Root().AllComponents()
		for val, err := allComponents.Next(); err == nil; val, err = allComponents.Next() {
			hCommon.Try(func() {
				if component, ok := val.(IBDestroy); ok {
					fmt.Println(";;;;;;;;;;;;;;;;;;")
					component.CheckClose(this.waitGroup)
				}
			})
		}
		this.waitGroup.Wait()

		hLog.Info("===== 所有组件服务完毕, 5秒后关闭服务器 =====")
		go func() {
			// 断开Master连接后 继续服务20秒
			// 用于无状态组件更新
			time.Sleep(time.Second * 5)
			this.Close <- struct{}{}
		}()

		select {
		case <-this.Close:
		}
	}

	client, err := this.GetMasterClient()

	if err == nil {
		var reply string
		_ = client.Call("MasterService.GetNodeStatus", hConfig.Config.ClusterConfig.LocalAddress, &reply)

		if reply == hCluster.NODE_STAUTE_WEBCLOSE {
			_ = client.CallWithoutReply("MasterService.SetNodeStatus", []string{hConfig.Config.ClusterConfig.LocalAddress, hCluster.NODE_STAUTE_WAITPM2CLOSE})
		}
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
		this.Config.ClusterConfig.WorkName = nodeConfName
		this.Config.ClusterConfig.LocalAddress = s.LocalAddress
		this.Config.ClusterConfig.Role = s.Role
		if s.NetAddr.Alias != "" {
			this.Config.ClusterConfig.NetListenAddressAlias = s.NetAddr.Alias
			this.Config.ClusterConfig.NetListenAddress = s.NetAddr.Addr
		} else if s.NetAddr.Addr != "" {
			this.Config.ClusterConfig.NetListenAddress = s.NetAddr.Addr
			this.Config.ClusterConfig.NetListenAddressAlias = "ws://" + s.NetAddr.Addr + "/ws"
		}
		this.Config.ClusterConfig.WorkId = s.WorkId
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

func (this *LauncherComponent) GetMasterClient() (*rpc.TcpClient, error) {
	if this.master == nil {
		var err error
		this.master, err = this.nodeComponent.GetNodeClient(hConfig.Config.ClusterConfig.MasterAddress)

		if err != nil {
			return nil, err
		}
	}
	return this.master, nil
}

func (this *LauncherComponent) CheckNodeStatus() {
	client, err := this.GetMasterClient()

	if err == nil {
		var reply string
		_ = client.Call("MasterService.GetNodeStatus", hConfig.Config.ClusterConfig.LocalAddress, &reply)
		if reply == hCluster.NODE_STAUTE_WAITPM2CLOSE {
			// 阻塞
			<-this.Close
		}
	}
}

//func (this *LauncherComponent) RegisterGateCheckCloseFunc(handler func()) {
//	this.checkHandler = handler
//}
