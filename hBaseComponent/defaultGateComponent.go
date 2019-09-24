package hBaseComponent

import (
	"custom/happy/hCluster"
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"custom/happy/hNet"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type DefaultGateComponent struct {
	hEcs.ComponentBase
	locker        sync.RWMutex
	nodeComponent *hCluster.NodeComponent
	//launcherComponent *LauncherComponent
	clients sync.Map // [sessionID,*session]
	NetAPI  hNet.ILogicAPI
	server  *hNet.Server

	acceptNum     int32
	isCloseServer bool
	waitGroup     *sync.WaitGroup
}

func (this *DefaultGateComponent) IsUnique() int {
	return hEcs.UNIQUE_TYPE_GLOBAL
}

func (this *DefaultGateComponent) GetRequire() map[*hEcs.Object][]reflect.Type {
	requires := make(map[*hEcs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&hConfig.ConfigComponent{}),
	}
	return requires
}

func (this *DefaultGateComponent) Awake(ctx *hEcs.Context) {
	this.isCloseServer = false
	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}

	//err = this.Parent().Root().Find(&this.launcherComponent)
	//if err != nil {
	//	panic(err)
	//}

	if this.NetAPI == nil {
		panic(errors.New("NetAPI is necessity of defaultGateComponent"))
	}

	this.NetAPI.Init(this.Parent())

	conf := &hNet.ServerConf{
		Protocol:             "ws",
		PackageProtocol:      &hNet.TdProtocol{},
		Address:              hConfig.Config.ClusterConfig.NetListenAddress,
		IsUsePool:            true,
		QueueCap:             10000,
		ReadTimeout:          time.Millisecond * time.Duration(hConfig.Config.ClusterConfig.NetConnTimeout),
		OnClientDisconnected: this.OnDropped,
		OnClientConnected:    this.OnConnected,
		LogicAPI:             this.NetAPI,
		MaxInvoke:            1000,
	}

	this.server = hNet.NewServer(conf)
	err = this.server.StartUp()
	if err != nil {
		panic(err)
	}

	//
	//this.launcherComponent.RegisterGateCheckCloseFunc(this.CheckClose)
}

func (this *DefaultGateComponent) CheckClose(group *sync.WaitGroup) {
	hLog.Info("Gate 网关检查关闭")
	//this.server.CheckClose()
	this.locker.RLock()
	defer this.locker.RUnlock()
	if atomic.LoadInt32(&this.acceptNum) > 0 {
		fmt.Println("Gate  ============= CheckClose No")
		group.Add(1)
	}
	this.isCloseServer = true
	this.waitGroup = group
}

func (this *DefaultGateComponent) AddNetAPI(api hNet.ILogicAPI) {
	this.NetAPI = api
}

func (this *DefaultGateComponent) OnConnected(sess *hNet.Session) {
	atomic.AddInt32(&this.acceptNum, 1)
	this.clients.Store(sess.Id, sess)
	this.NetAPI.OnConnect(sess)
	hLog.Debug(fmt.Sprintf("client [ %s ] connected,session id :[ %s ]", sess.RemoteAddr(), sess.Id))
}

func (this *DefaultGateComponent) OnDropped(sess *hNet.Session) {
	defer func() {
		this.locker.RLock()
		if this.isCloseServer && atomic.LoadInt32(&this.acceptNum) <= 0 && this.waitGroup != nil {
			this.waitGroup.Done()
		}
		this.locker.RUnlock()
	}()

	atomic.AddInt32(&this.acceptNum, -1)
	this.clients.Delete(sess.Id)
	this.NetAPI.OnDisconnect(sess)
}

func (this *DefaultGateComponent) Destroy(ctx *hEcs.Context) {
	hLog.Info("服务器关闭  Shutdown")
	this.server.Shutdown()
}

func (this *DefaultGateComponent) SendMessage(sid string, message interface{}) error {
	if s, ok := this.clients.Load(sid); ok {
		this.NetAPI.Reply(s.(*hNet.Session), message)
	}
	return errors.New(fmt.Sprintf("this session id: [ %s ] not exist", sid))
}

func (this *DefaultGateComponent) Emit(sess *hNet.Session, message interface{}) {
	this.NetAPI.Reply(sess, message)
}
