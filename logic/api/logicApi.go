package api

import (
	"../../hActor"
	"../../hCluster"
	"../../hNet"
	"../../hNet/messageProtocol"
	"../../hLog"
	"../logicComponents"
	"fmt"
)

type LogicApi struct {
	hNet.ApiBase
	nodeComponent *hCluster.NodeComponent
	actorProxy *hActor.ActorProxyComponent
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
		Msg: "Hello Client: 我收到了你的消息",
	})
}

func (this *LogicApi) Login(sess *hNet.Session, message *LoginMessage) {
	errReply := func(msg string) {
		r := &CommonResMessage{
			Statue: CODE_ERROR,
			Msg: msg,
		}
		this.Reply(sess, r)
	}

	serviceCaller, err := this.Upgrade(sess)
	if err != nil {
		errReply("服务器Login回话转换失败")
		return
	}

	reply, err := serviceCaller.Call("login", logicComponents.Service_Login_Login, message.Nickname)
	if err != nil {
		hLog.Debug(err);
		errReply("登录失败服务器登录节点异常")
		return
	}

	this.Reply(sess, &LoginResMessage{
		Statue: CODE_OK,
		Msg: reply[0].(string),
	})
}

