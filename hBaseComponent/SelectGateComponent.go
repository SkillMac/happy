package hBaseComponent

import (
	"custom/happy/hCluster"
	"custom/happy/hCommon"
	"custom/happy/hConfig"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"custom/happy/hRpc"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

/**
这组件是构建无状态的路由网关组件
由它决定到底路由到那个网关上面
*/

type SelectGateComponent struct {
	hEcs.ComponentBase
	nodeComponent *hCluster.NodeComponent
	router        *gin.Engine
	master        *rpc.TcpClient
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

func (this *SelectGateComponent) initServer() {
	gin.SetMode(gin.ReleaseMode)
	this.router = gin.New()
	this.router.Use(gin.Recovery())
	this.router.Use(hCommon.Cors())
	this.router.GET("/", this.home)
	//this.router.POST("/getUsableGate", this.getUsableGate)
	this.router.POST("/getUsableGate", this.getUsableGate)
	this.router.POST("/getOnLineWork", this.getOnLineWork)
	this.router.POST("/closeOnLineNode", this.closeOnLineNode)
	this.router.POST("/getAllNodeStatus", this.getAllNodeStatus)
	this.router.POST("/setNodeStatus", this.setNodeStatus)

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
		return
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

func (this *SelectGateComponent) getOnLineWork(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer

	if r.Method != "POST" {
		http.Error(w, "Post type is not POST", http.StatusMethodNotAllowed)
		return
	}

	master, err := this.GetMasterClient()
	if err != nil {
		_, err = w.WriteString("{\"err\":\"[ERROR] ==>" + err.Error() + "\"}")
		if err != nil {
			hLog.Error("No Status Gate {getOnLineWork} WriteString fail, ERROR => ", err)
		}
		return
	}

	var reply *hCluster.NodeInfoSyncReply
	err = master.Call("MasterService.NodeInfoSync", "sync", &reply)

	if err != nil {
		_, err = w.WriteString("{\"err\":\"[ERROR] ==>" + err.Error() + "\"}")
		if err != nil {
			hLog.Error("No Status Gate {getOnLineWork} WriteString fail, ERROR => ", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Nodes": reply.Nodes})
}

func (this *SelectGateComponent) closeOnLineNode(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer

	if r.Method != "POST" {
		http.Error(w, "Post type is not POST", http.StatusMethodNotAllowed)
		return
	}

	errReply := func(msg string) {
		ctx.JSON(http.StatusOK, gin.H{"msg": msg})
		hLog.Error("No Status Gate {closeOnLineNode} fail ==>", msg)
	}

	name := r.PostFormValue("name") //string(b[0:n])
	hLog.Info("[No Gate] Close Server   ", name)
	if cfg, ok := hConfig.Config.ClusterConfig.NodeDefine[name]; ok {
		// 通知 master 改变当前节点的状态
		client, err := this.GetMasterClient()

		if err != nil {
			errReply(err.Error())
			return
		}

		err = client.CallWithoutReply("MasterService.SetNodeStatus", []string{cfg.LocalAddress, hCluster.NODE_STAUTE_WEBCLOSE})
		if err != nil {
			errReply(err.Error())
			return
		}

		client, err = this.nodeComponent.GetNodeClient(cfg.LocalAddress)
		if err != nil {
			errReply(err.Error())
			return
		}
		err = client.CallWithoutReply("LauncherService.CloseSelf", name)

		if err != nil {
			errReply(err.Error())
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"msg": "ok"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"msg": "no exists"})
	}

}

func (this *SelectGateComponent) getAllNodeStatus(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer

	if r.Method != "POST" {
		http.Error(w, "Post type is not POST", http.StatusMethodNotAllowed)
		return
	}

	errReply := func(msg string) {
		ctx.JSON(http.StatusOK, gin.H{"msg": msg})
		hLog.Error("No Status Gate {getAllNodeStatus} fail ==>", msg)
	}

	master, err := this.GetMasterClient()

	if err != nil {
		errReply(err.Error())
		return
	}
	var reply map[string]string
	err = master.Call("MasterService.GetAllNodeStatus", "", &reply)
	if err != nil {
		errReply(err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "ok", "AllNodeStatus": reply})
}

func (this *SelectGateComponent) setNodeStatus(ctx *gin.Context) {
	r := ctx.Request
	w := ctx.Writer

	if r.Method != "POST" {
		http.Error(w, "Post type is not POST", http.StatusMethodNotAllowed)
		return
	}
	name := r.PostFormValue("name") //string(b[0:n])

	errReply := func(msg string) {
		ctx.JSON(http.StatusOK, gin.H{"msg": msg})
		hLog.Error("No Status Gate {setNodeStatus} fail ==>", msg)
	}

	master, err := this.GetMasterClient()

	if err != nil {
		errReply(err.Error())
		return
	}
	err = master.CallWithoutReply("MasterService.SetNodeStatus", []string{name, hCluster.NODE_STAUTE_OFFLINE})
	if err != nil {
		errReply(err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (this *SelectGateComponent) GetMasterClient() (*rpc.TcpClient, error) {
	if this.master == nil {
		var err error
		this.master, err = this.nodeComponent.GetNodeClient(hConfig.Config.ClusterConfig.MasterAddress)

		if err != nil {
			return nil, err
		}
	}
	return this.master, nil
}
