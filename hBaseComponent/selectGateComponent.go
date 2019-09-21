package hBaseComponent

import (
	"custom/happy/hCluster"
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strings"
)

/**
这组件是构建无状态的路由网关组件
由它决定到底路由到那个网关上面
*/

type SelectGateComponent struct {
	hEcs.ComponentBase
	nodeComponent *hCluster.NodeComponent
	router        *gin.Engine
}

func (this *SelectGateComponent) IsUnique() int {
	return hEcs.UNIQUE_TYPE_GLOBAL
}

func (this *SelectGateComponent) GetRequire() map[*hEcs.Object][]reflect.Type {
	requires := make(map[*hEcs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&hConfig.ConfigComponent{}),
	}
	return requires
}

func (this *SelectGateComponent) Awake(ctx *hEcs.Context) {
	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}
	this.initServer()
	this.StartUp()
}

////// 跨域
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               //请求方法
		origin := c.Request.Header.Get("Origin") //请求头部
		var headerKeys []string                  // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")                                       // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")                                                                                                                                                              // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next() //  处理请求
	}
}

func (this *SelectGateComponent) initServer() {
	gin.SetMode(gin.ReleaseMode)
	this.router = gin.New()
	this.router.Use(gin.Recovery())
	this.router.Use(Cors())
	this.router.GET("/", this.home)
	//this.router.POST("/getUsableGate", this.getUsableGate)
	this.router.POST("/getUsableGate", this.getUsableGate)

}

func (this *SelectGateComponent) StartUp() {
	if this.router == nil {
		return
	}

	go func() {
		hLog.Info("No Status Net Run Successful, Listener ==> ", hConfig.Config.ClusterConfig.SelectNetListenAddress)
		err := this.router.Run(hConfig.Config.ClusterConfig.SelectNetListenAddress)
		if err != nil {
			hLog.Error("No Status Net Run Fail, Err ==> ")
			panic(err)
			return
		}
	}()
}

func (this *SelectGateComponent) home(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer
	hLog.Debug("Url ==> home")
	if r.URL.Path != "/" {
		http.Error(w, "Api not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Post type is not GET", http.StatusMethodNotAllowed)
	}

	_, err := w.WriteString("happy server framework")
	if err != nil {
		hLog.Error("No Status Gate {home} WriteString fail, ERROR => ", err)
	}
}

func (this *SelectGateComponent) getUsableGate(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer

	if r.Method != "POST" {
		http.Error(w, "Post type is not POST", http.StatusMethodNotAllowed)
	}
	nodeIdGroup, err := this.nodeComponent.GetNodeGroup("gate")
	if err != nil {
		hLog.Error("err ", err)
		_, err := w.WriteString("{\"err\":\"[ERROR] ==>" + err.Error() + "\"}")
		if err != nil {
			hLog.Error("No Status Gate {getUsableGate} WriteString fail, ERROR => ", err)
		}
		return
	}

	nodeInfoDetail, err := nodeIdGroup.SelectOneNodeInfo(hCluster.SELECTOR_TYPE_MIN_LOAD)

	if err != nil {
		hLog.Error("err", err)
		_, err := w.WriteString("{\"err\":\"[ERROR] ==>" + err.Error() + "\"}")
		if err != nil {
			hLog.Error("No Status Gate {getUsableGate} WriteString fail, ERROR => ", err)
		}
	}
	fmt.Printf("%v  =========", nodeInfoDetail.CustomData)
	_, err = w.WriteString("{\"ip\":\"" + nodeInfoDetail.CustomData["netAddr"].(string) + "\"}")
	if err != nil {
		hLog.Error("No Status Gate {getUsableGate} WriteString fail, ERROR => ", err)
	}
}
