package api

import (
	"../../hActor"
	"../../hCluster"
	"../../hLog"
	"../../hNet"
	"../../hNet/messageProtocol"
	"../components"
	"container/list"
	"fmt"
	"sync"
	"time"
)

type LogicApi struct {
	hNet.ApiBase
	nodeComponent   *hCluster.NodeComponent
	actorProxy      *hActor.ActorProxyComponent
	matchSessionMap list.List // *innerMatchPlayer
	rwLock          sync.RWMutex
}

func NewLogicApi() *LogicApi {
	ta := &LogicApi{}
	ta.Instance(ta).SetMT2ID(Id2mt).SetProtocol(&messageProtocol.JsonProtocol{})
	return ta
}

/*
提供给本脚本使用的方法
**/
func (this *LogicApi) ActorProxy() (*hActor.ActorProxyComponent, error) {
	if this.actorProxy == nil {
		p, err := this.GetParent()
		if err != nil {
			return nil, err
		}
		err = p.Root().Find(&this.actorProxy)
		if err != nil {
			return nil, err
		}
	}
	return this.actorProxy, nil
}

func (this *LogicApi) Upgrade(sess *hNet.Session) (*hActor.ActorServiceCaller, error) {
	proxy, err := this.ActorProxy()
	if err != nil {
		return nil, err
	}
	return hActor.NewActorServiceCallerFromSession(sess, proxy), nil
}

func (this *LogicApi) NodeComponent() (*hCluster.NodeComponent, error) {
	if this.nodeComponent == nil {
		o, err := this.GetParent()
		if err != nil {
			return nil, err
		}
		err = o.Root().Find(&this.nodeComponent)
		if err != nil {
			return nil, err
		}
		return this.nodeComponent, nil
	}
	return this.nodeComponent, nil
}

/*
路由的方法
这个是gate网关服务器的逻辑 他决定你的逻辑会在那个 逻辑节点上处理
**/
func (this *LogicApi) Hello(sess *hNet.Session, message *TestMessage) {
	println(fmt.Sprintf("hello %s", message.Name))
	this.Reply(sess, &CommonResMessage{
		Statue: CODE_OK,
		Msg:    "Hello Client: 我收到了你的消息",
	})
}

func (this *LogicApi) Login(sess *hNet.Session, message *LoginMessage) {
	errReply := func(msg string) {
		r := &CommonResMessage{
			Statue: CODE_ERROR,
			Msg:    msg,
		}
		this.Reply(sess, r)
	}

	serviceCaller, err := this.Upgrade(sess)
	if err != nil {
		errReply("服务器Login回话转换失败")
		return
	}

	reply, err := serviceCaller.Call("login", components.Service_Login_Login, message.Nickname)
	if err != nil {
		hLog.Debug(err)
		errReply("登录失败服务器登录节点异常")
		return
	}

	this.Reply(sess, &LoginResMessage{
		Statue: CODE_OK,
		Msg:    reply[0].(string),
	})
}

/**
匹配相关的逻辑
*/

type innerMatchPlayer struct {
	sid     string
	lv      int
	session *hNet.Session
}

func (this *LogicApi) MatchTimer() {
	timer := time.NewTicker(time.Second)

	// 每秒轮训查找一次匹配列表中
	for {
		select {
		case <-timer.C:
			{

			}
		}
	}
}

func (this *LogicApi) Match(session *hNet.Session, message *MatchMessage) {
	//r := &MatchResMessage{
	//	CommonResMessage{
	//		Statue: CODE_OK,
	//		Msg:    "",
	//	}, components.MatchPlayInfo{
	//		NickName: "",
	//		HeadUrl:  "",
	//		Lv:       0,
	//	},
	//}
	//
	//session.SetProperty("lv", 1)
	//
	//errReply := func(msg string) {
	//	r.Statue = CODE_ERROR
	//	r.Msg = msg
	//	this.Reply(session, r)
	//}

	//lv, ok := session.GetProperty("lv")
	//if !ok {
	//	hLog.Error("网关服务器 匹配 获取玩家等级异常")
	//	errReply("网关服务器 匹配 获取玩家等级异常")
	//	return
	//}

}
