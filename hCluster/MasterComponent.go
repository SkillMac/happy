package hCluster

import (
	"custom/happy/hCommon"
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	LOG_TYPE_NODE_CLOSE = iota
)

type MasterComponent struct {
	hEcs.ComponentBase
	locker          *sync.RWMutex
	nodeComponent   *NodeComponent
	Nodes           map[string]*NodeInfo
	NodeLog         *NodeLogs
	timeoutChecking map[string]int
}

func (this *MasterComponent) GetRequire() map[*hEcs.Object][]reflect.Type {
	requires := make(map[*hEcs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&hConfig.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *MasterComponent) Awake(ctx *hEcs.Context) {
	this.locker = &sync.RWMutex{}
	this.Nodes = make(map[string]*NodeInfo)
	this.NodeLog = &NodeLogs{BufferSize: 20}
	this.timeoutChecking = make(map[string]int)

	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}

	//注册Master服务
	s := new(MasterService)
	s.init(this)
	err = this.nodeComponent.Register(s)
	if err != nil {
		panic(err)
	}
	if !hConfig.Config.CommonConfig.Debug || false {
		go this.TimeoutCheck()
	}
}

//上报节点信息
func (this *MasterComponent) UpdateNodeInfo(args *NodeInfo) {
	this.locker.Lock()
	if _, ok := this.Nodes[args.Address]; !ok {
		s := strings.Builder{}
		for _, value := range args.Role {
			s.WriteString(value)
			s.WriteString("  ")
		}
		hLog.Info(fmt.Sprintf("Node [ %s ] connected to this master, roles: [ %s]", args.Address, s.String()))
	}
	args.Time = time.Now().UnixNano()
	this.Nodes[args.Address] = args
	this.timeoutChecking[args.Address] = 0

	this.locker.Unlock()
}

//节点主动关闭
func (this *MasterComponent) NodeClose(addr string) {
	//非线程安全，外层注意加锁
	if v, ok := this.Nodes[addr]; ok {
		s := strings.Builder{}
		for _, value := range v.Role {
			s.WriteString(value)
			s.WriteString("  ")
		}
		hLog.Info(fmt.Sprintf("Node [ %s ] disconnected, roles: [ %s ]", addr, s.String()))
		//if !hConfig.Config.CommonConfig.Debug {
		//	body := fmt.Sprintf(`IP [ %s ] 断开连接<br>
		//	Role [ %s ]`, addr, s.String())
		//	hCommon.SendEmail(fmt.Sprintf("节点断开连接 [ %s]", s.String()), body)
		//}
	}
	delete(this.Nodes, addr)
	delete(this.timeoutChecking, addr)
	this.NodeLog.Add(&NodeLog{
		Time: time.Now().UnixNano(),
		Type: LOG_TYPE_NODE_CLOSE,
		Log:  addr,
	})
}

//查询节点信息 args : "AppID:Role:SelectorType"
func (this *MasterComponent) NodeInquiry(args []string, detail bool) ([]*InquiryReply, error) {
	return Selector(this.Nodes).DoQuery(args, detail, this.locker)
}

//检查超时节点
func (this *MasterComponent) TimeoutCheck() map[string]*NodeInfo {
	var interval = time.Duration(hConfig.Config.ClusterConfig.ReportInterval)
	for {
		time.Sleep(time.Millisecond * interval)
		this.locker.Lock()
		for addr, count := range this.timeoutChecking {
			this.timeoutChecking[addr] = count + 1
			if this.timeoutChecking[addr] > 3 {
				this.NodeClose(addr)
			}
		}
		this.locker.Unlock()
	}
}

//深度复制节点信息
func (this *MasterComponent) NodesCopy() map[string]*NodeInfo {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return hCommon.Copy(this.Nodes).(map[string]*NodeInfo)
}

func (this *MasterComponent) NodesLogsCopy() *NodeLogs {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return hCommon.Copy(this.NodeLog).(*NodeLogs)
}
