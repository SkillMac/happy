package hNet

import (
	"context"
	"custom/happy/hCommon"
	"custom/happy/hConfig"
	"custom/happy/hLog"
	"encoding/binary"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

type HttpConn struct {
	locker   sync.Mutex
	httpConn *gin.Context
}

func (this *HttpConn) Addr() string {
	return this.httpConn.ClientIP()
}

func (this *HttpConn) WriteMessage(messageType uint32, data []byte) error {
	msg := make([]byte, 4)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], messageType)
	this.locker.Lock()
	defer this.locker.Unlock()
	this.httpConn.JSON(http.StatusOK, gin.H{"Data": msg})
	return nil
}

func (this *HttpConn) Close() error {
	return nil
}

type MsgData struct {
	Data []byte
}

type HttpHandler struct {
	conf      *ServerConf
	server    *Server
	acceptNum int32
	gpool     *gPool
}

func (this *HttpHandler) Listen() error {
	conf := this.conf
	if conf.IsUsePool && conf.MaxInvoke == 0 {
		conf.MaxInvoke = 20
	}
	this.gpool = GetGloblePool(int(conf.MaxInvoke), conf.QueueCap)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(hCommon.Cors())
	router.POST("/api", this.Api)

	go func() {
		err := router.Run(hConfig.Config.ClusterConfig.NetListenAddress)
		if err != nil {
			hLog.Critical(err)
		}
	}()
	return nil
}

func (this *HttpHandler) Api(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer

	if r.Method != "POST" {
		http.Error(w, "Post type is not POST", http.StatusMethodNotAllowed)
		return
	}

	//atomic.AddInt32(&this.acceptNum, 1)
	sess := NewSeesion(hCommon.GetUUID(), &HttpConn{httpConn: ctx})

	var p MsgData
	err := ctx.BindJSON(&p)
	if err != nil {

	} else {
		this.revc(sess, &p)
	}
}

func (this *HttpHandler) revc(sess *Session, p *MsgData) {
	sess.SetProperty("workerId", int32((-1)))
	this.handleMsg(sess, p.Data)

	wid := int32(-1)

	if this.conf.IsUsePool {
		// TODO
		this.gpool.AddJobSerial(this.handleMsg, []interface{}{sess, p, sess.Id}, wid, func(workerId int32) {
			wid = workerId
			sess.SetProperty("workerId", wid)
		})
	} else {
		go this.handleMsg(sess, p)
	}
}

func (this *HttpHandler) handleMsg(args ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "cid", args[0])
	if this.conf.CustomeHandler != nil {
		this.conf.CustomeHandler(args[0].(*Session), args[1].([]byte))
	} else {
		mid, mes := this.conf.PackageProtocol.Decode(ctx, args[1].([]byte))
		if this.conf.LogicAPI != nil && mid != nil {
			this.server.invoke(ctx, mid[0], mes)
		} else {
			hLog.Error("[Error] no message handler")
		}
	}
}

func (this *HttpHandler) Handle() error {
	return nil
}

func (this *HttpHandler) CheckClose() {

}

func (this *HttpHandler) Destroy() error {
	if this.gpool != nil {
		hLog.Info("HttpHandler 协程池释放")
		this.gpool.Release()
	}
	return nil
}
