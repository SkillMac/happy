package hCluster

import (
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hRpc"
	"errors"
	"reflect"
	"sync"
	"time"
)

/*
* 可能会因为更新不及时 导致 连接失败,  暂时不用
 */
type LocationReply struct {
	NodeNetAddress map[string]string //[node id , ip]
}
type LocationQuery struct {
	Group  string
	AppID  string
	NodeID string
}

type LocationComponent struct {
	hEcs.ComponentBase
	locker        *sync.RWMutex
	nodeComponent *NodeComponent
	Nodes         map[string]*NodeInfo
	NodeLog       *NodeLogs
	master        *rpc.TcpClient
}

func (this *LocationComponent) GetRequire() map[*hEcs.Object][]reflect.Type {
	requires := make(map[*hEcs.Object][]reflect.Type)
	requires[this.Runtime().Root()] = []reflect.Type{
		reflect.TypeOf(&hConfig.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *LocationComponent) Awake(ctx *hEcs.Context) {
	this.locker = &sync.RWMutex{}
	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}

	//注册位置服务节点RPC服务
	service := new(LocationService)
	service.init(this)
	err = this.nodeComponent.Register(service)
	if err != nil {
		panic(err)
	}
	go this.DoLocationSync()
}

//同步节点信息到位置服务组件
func (this *LocationComponent) DoLocationSync() {
	var reply *NodeInfoSyncReply
	var interval = time.Duration(hConfig.Config.ClusterConfig.LocationSyncInterval)
	for {
		if this.master == nil {
			var err error
			this.master, err = this.nodeComponent.GetNodeClient(hConfig.Config.ClusterConfig.MasterAddress)
			if err != nil {
				time.Sleep(time.Second * interval)
				continue
			}
		}
		err := this.master.Call("MasterService.NodeInfoSync", "sync", &reply)
		if err != nil {
			this.master = nil
			continue
		}

		this.locker.Lock()
		this.UpdateBL(this.Nodes, reply.Nodes)
		this.Nodes = reply.Nodes
		this.NodeLog = reply.NodeLog
		this.locker.Unlock()
		time.Sleep(time.Millisecond * interval)
	}
}

//查询节点信息 args : "AppID:Role:SelectorType"
func (this *LocationComponent) NodeInquiry(args []string, detail bool) ([]*InquiryReply, error) {
	if this.Nodes == nil {
		return nil, errors.New("this location node is waiting to sync")
	}
	return Selector(this.Nodes).DoQuery(args, detail, this.locker)
}

//日志获取
func (this *LocationComponent) NodeLogInquiry(args int64) ([]*NodeLog, error) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	if this.NodeLog == nil {
		return nil, errors.New("this location node is waiting to sync")
	}
	return this.NodeLog.Get(args), nil
}

func (this *LocationComponent) UpdateBL(curNodes map[string]*NodeInfo, nextNodes map[string]*NodeInfo) {
	if curNodes == nil {
		// 直接添加
		for addr, _ := range nextNodes {
			this.NodeOpen(addr)
		}
		return
	}
	for addr, _ := range curNodes {
		if _, ok := nextNodes[addr]; !ok {
			this.NodeClose(addr)
		}
	}

	for addr, _ := range nextNodes {
		if _, ok := curNodes[addr]; !ok {
			this.NodeOpen(addr)
		}
	}
}

func (this *LocationComponent) NodeOpen(addr string) {
	r2LB.Add(addr)
	p2cLB.Add(addr)
	boundedLB.Add(addr)
}

func (this *LocationComponent) NodeClose(addr string) {
	r2LB.Remove(addr)
	p2cLB.Remove(addr)
	boundedLB.Remove(addr)
}
