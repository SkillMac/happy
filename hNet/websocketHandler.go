package hNet

import (
	"context"
	"custom/happy/hCommon"
	"custom/happy/hLog"
	"encoding/binary"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var upgrader = websocket.Upgrader{
	//HandshakeTimeout:  0,
	//ReadBufferSize:    0,
	//WriteBufferSize:   0,
	//WriteBufferPool:   nil,
	//Subprotocols:      nil,
	//Error:             nil,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	//EnableCompression: false,
}

type WsConn struct {
	locker sync.Mutex
	wsConn *websocket.Conn
}

func (this *WsConn) Addr() string {
	return this.wsConn.RemoteAddr().String()
}

func (this *WsConn) WriteMessage(messageType uint32, data []byte) error {
	msg := make([]byte, 4)
	msg = append(msg, data...)
	binary.BigEndian.PutUint32(msg[:4], messageType)
	this.locker.Lock()
	defer this.locker.Unlock()
	return this.wsConn.WriteMessage(2, msg)
}

func (this *WsConn) Close() error {
	return this.wsConn.Close()
}

/**
implent IHandler interface
*/
type WebSocketHandler struct {
	conf      *ServerConf
	server    *Server
	numInvoke int32
	acceptNum int32
	invokeNum int32
	idleTime  time.Time
	gpool     *gPool
}

func (this *WebSocketHandler) Listen() error {
	conf := this.conf
	if conf.IsUsePool && conf.MaxInvoke == 0 {
		conf.MaxInvoke = 20
	}
	this.gpool = GetGloblePool(int(conf.MaxInvoke), conf.QueueCap)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/", home)
	router.GET("/getAddress", getUsableGateAddress)
	router.GET("/ws", func(ctx *gin.Context) {
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			_, _ = ctx.Writer.WriteString("Server internal error")
			return
		}
		log.Println("Ws accept:", conn.RemoteAddr())
		atomic.AddInt32(&this.acceptNum, 1)
		sess := NewSeesion(hCommon.GetUUID(), &WsConn{wsConn: conn})
		if conf.OnClientConnected != nil {
			conf.OnClientConnected(sess)
		}
		this.recv(sess, conn)
		sess.rwLock.Lock()
		sess.conn = nil
		sess.rwLock.Unlock()
		if conf.OnClientDisconnected != nil {
			conf.OnClientDisconnected(sess)
		}
		atomic.AddInt32(&this.acceptNum, -1)
	})

	go func() {
		fmt.Println(fmt.Sprintf("Websocket server listening port [ %s ]", conf.Address))
		err := router.Run(conf.Address)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()
	return nil
}

func (this *WebSocketHandler) Handle() error {
	return nil
}

func (this *WebSocketHandler) recv(sess *Session, conn *websocket.Conn) {
	defer conn.Close()

	sess.SetProperty("workerId", int32((-1)))
	handler := func(args ...interface{}) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "cid", args[0])
		if this.conf.CustomeHandler != nil {
			this.conf.CustomeHandler(args[0].(*Session), args[1].([]byte))
		} else {
			mid, mes := this.conf.PackageProtocol.Decode(ctx, args[1].([]byte))
			if this.conf.LogicAPI != nil && mid != nil {
				this.server.invoke(ctx, mid[0], mes)
			} else {
				fmt.Println("[Error] no message handler")
			}
		}
	}

	wid := int32(-1)
	for !this.server.isClosed {
		_, pkg, err := conn.ReadMessage()
		if err != nil || pkg == nil {
			hLog.Warn("[Error]", fmt.Sprintf("Close connection %s: %v", this.conf.Address, err))
			return
		}
		if this.conf.IsUsePool {
			// TODO
			this.gpool.AddJobSerial(handler, []interface{}{sess, pkg, sess.Id}, wid, func(workerId int32) {
				wid = workerId
			})
		} else {
			go handler(sess, pkg)
		}
	}
}

func (this *WebSocketHandler) Destroy() error {
	if this.gpool != nil {
		hLog.Info("协程池释放")
		this.gpool.Release()
	}
	return nil
}

/**
New
*/
func NewWebSocketHandler(conf *ServerConf, server *Server) *WebSocketHandler {
	wsh := &WebSocketHandler{conf: conf, server: server}
	return wsh
}

func home(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer
	log.Print(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Api not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Post type is not GET", http.StatusMethodNotAllowed)
	}

	//http.ServeFile(w, r, "home.html")
	w.WriteString("happy server framework")
}

func getUsableGateAddress(ctx *gin.Context) {

}
