package hCluster

import (
	"custom/happy/hCommon"
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"custom/happy/hRpc"
	"fmt"
	"github.com/struCoder/pidusage"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

type ChildComponent struct {
	hEcs.ComponentBase
	locker          sync.RWMutex
	localAddr       string
	netAddr         string
	rpcMaster       *rpc.TcpClient //master节点
	nodeComponent   *NodeComponent
	sysInfo         *pidusage.SysInfo
	reportCollecter func() map[string]float64
	close           bool
}

func (this *ChildComponent) GetRequire() map[*hEcs.Object][]reflect.Type {
	requires := make(map[*hEcs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&hConfig.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *ChildComponent) Awake(ctx *hEcs.Context) {
	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}
	// 添加 cpu 和 内存收集器
	this.AddReportInfo("custom", func() map[string]float64 {
		this.sysInfo, err = pidusage.GetStat(os.Getpid())
		if err != nil {
			hLog.Error(fmt.Sprintf("[Role]: %v, [Addr]: %s 获取系统信息失败", hConfig.Config.ClusterConfig.Role, this.localAddr))
			return make(map[string]float64)
		}
		return map[string]float64{
			"cpu": this.sysInfo.CPU,
			"mem": this.sysInfo.Memory,
		}
	})
	go this.ConnectToMaster()
	go this.DoReport()
}

func (this *ChildComponent) Destroy(ctx *hEcs.Context) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.close = true
	this.ReportClose(this.localAddr)
}

//上报节点信息
func (this *ChildComponent) DoReport() {
	hCommon.When(time.Millisecond*50,
		func() bool {
			this.locker.RLock()
			defer this.locker.RUnlock()

			return this.rpcMaster != nil
		},
		func() bool {
			this.locker.RLock()
			defer this.locker.RUnlock()
			return this.localAddr != ""
		})
	args := &NodeInfo{
		Address:    this.localAddr,
		Role:       hConfig.Config.ClusterConfig.Role,
		AppName:    hConfig.Config.ClusterConfig.AppName,
		CustomData: map[string]interface{}{"netAddr": this.netAddr},
	}
	var reply bool
	var interval = time.Duration(hConfig.Config.ClusterConfig.ReportInterval)
	for {
		reply = false
		this.locker.RLock()
		if this.close {
			this.locker.RUnlock()
			return
		}
		m := make(map[string]float64)
		//for _, collector := range this.reportCollecter {
		//	f, d := collector()
		//	m[f] = d
		//}
		if this.reportCollecter != nil {
			m = this.reportCollecter()
			//hLog.Debug("上传节点信息", m)
		}
		args.Info = m
		this.locker.RUnlock()
		if this.rpcMaster != nil {
			err := this.rpcMaster.Call("MasterService.ReportNodeInfo", args, &reply)
			if err != nil {
				hLog.Error("Call MasterServie.ReportNodeInfo fail In ChildComponet")
			}
		}
		time.Sleep(time.Millisecond * interval)
	}
}

//增加上报信息
func (this *ChildComponent) AddReportInfo(field string, collectFunction func() map[string]float64) {
	this.locker.Lock()
	this.reportCollecter = collectFunction
	this.locker.Unlock()
}

//增加上报节点关闭
func (this *ChildComponent) ReportClose(addr string) {
	var reply bool
	if this.rpcMaster != nil {
		_ = this.rpcMaster.Call("MasterService.ReportNodeClose", addr, &reply)
	}
}

//连接到master
func (this *ChildComponent) ConnectToMaster() {
	addr := hConfig.Config.ClusterConfig.MasterAddress
	callback := func(event string, data ...interface{}) {
		switch event {
		case "close":
			this.OnDropped()
		}
	}
	hLog.Info(" Looking for master ......")
	var err error
	for {
		this.locker.Lock()
		this.rpcMaster, err = this.nodeComponent.ConnectToNode(addr, callback)
		if err == nil {
			ip := strings.Split(this.rpcMaster.LocalAddr(), ":")[0]
			port := strings.Split(hConfig.Config.ClusterConfig.LocalAddress, ":")[1]
			this.localAddr = fmt.Sprintf("%s:%s", ip, port)
			if hCommon.Contains(hConfig.Config.ClusterConfig.Role, "gate") {
				this.netAddr = hConfig.Config.ClusterConfig.NetListenAddressAlias
			}
			this.locker.Unlock()
			break
		}
		this.locker.Unlock()
		time.Sleep(time.Millisecond * 500)
	}

	this.nodeComponent.Locker().Lock()
	this.nodeComponent.isOnline = true
	this.nodeComponent.Locker().Unlock()

	hLog.Info(fmt.Sprintf("Connected to master [ %s ]", addr))
}

//当节点掉线
func (this *ChildComponent) OnDropped() {
	//重新连接 time.Now().Format("2006-01-02T 15:04:05")
	this.locker.RLock()
	if this.close {
		this.locker.RUnlock()
		return
	}
	this.locker.RUnlock()
	this.nodeComponent.Locker().Lock()
	this.nodeComponent.isOnline = false
	this.nodeComponent.Locker().Unlock()

	hLog.Info("Disconnected from master")
	this.ConnectToMaster()
}
