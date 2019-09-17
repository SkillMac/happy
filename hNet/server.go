package hNet

import (
	"context"
	"sync/atomic"
	"time"
)

type ServerConf struct {
	Protocol        string // ws tcp udp
	PackageProtocol IPackageProtocol
	LogicAPI        ILogicAPI
	CustomeHandler  func(session *Session, data []byte) // 自定义消息处理
	Address         string
	IsUsePool       bool
	MaxInvoke       int32
	QueueCap        int

	// Tcp 相关配置 未实现
	TCPReadBuffer  int
	TCPWriteBuffer int
	TCPNoDelay     bool

	OnClientConnected    func(sess *Session)
	OnClientDisconnected func(sess *Session)

	AcceptTimeout time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
}

type Server struct {
	conf       *ServerConf
	lastInvoke time.Time // 上一次调用的时间
	idleTime   time.Time // 待机时间
	isClosed   bool      // 是否关闭
	numInvoke  int32     // 调用的次数
	handle     IHander   // 当前协议的句柄
}

/**
New
*/
func NewServer(conf *ServerConf) *Server {
	ns := &Server{conf: conf}
	ns.isClosed = false
	ns.lastInvoke = time.Now()
	return ns
}

func (this *Server) StartUp() error {
	h := this.getHandler()
	this.handle = h
	if err := h.Listen(); err != nil {
		return err
	}
	return h.Handle()
}

/**
private
*/
func (this *Server) getHandler() IHander {
	var h IHander = nil
	if this.conf.Protocol == "ws" {
		h = NewWebSocketHandler(this.conf, this)
	} else if this.conf.Protocol == "udp" {
		// TODO
		panic("unsupport protocol: " + this.conf.Protocol)
	} else if this.conf.Protocol == "tcp" {
		// TODO
		panic("unsupport protocol: " + this.conf.Protocol)
	} else {
		panic("unsupport protocol: " + this.conf.Protocol)
	}
	return h
}

func (this *Server) CheckClose() {
	if this.handle != nil {
		this.handle.CheckClose()
	}
}

func (this *Server) Shutdown() {
	this.isClosed = true
	if this.handle != nil {
		this.handle.Destroy()
		this.handle = nil
	}
}

func (this *Server) GetConfig() *ServerConf {
	return this.conf
}

func (this *Server) IsZomibe(timeout time.Duration) bool {
	conf := this.GetConfig()
	return conf.MaxInvoke != 0 && this.numInvoke == conf.MaxInvoke && this.lastInvoke.Add(timeout).Before(time.Now())
}

func (this *Server) invoke(ctx context.Context, mid uint32, data []byte) {
	atomic.AddInt32(&this.numInvoke, 1)
	if sess, ok := ctx.Value("cid").(*Session); ok {
		this.conf.LogicAPI.Route(sess, mid, data)
	}
	atomic.AddInt32(&this.numInvoke, -1)
}
