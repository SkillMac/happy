package api

import (
	"../../hActor"
	"../../hCluster"
	"../../hLog"
	"../../hNet"
	"../../hNet/messageProtocol"
	"../components"
	"fmt"
	"sync"
	"time"
)

type LogicApi struct {
	hNet.ApiBase
	nodeComponent   *hCluster.NodeComponent
	actorProxy      *hActor.ActorProxyComponent
	matchSessionMap map[string]*innerMatchPlayer
	chanMatchPlay   map[string]chan *innerMatchPlayer
	rwLock          sync.RWMutex
}

func NewLogicApi() *LogicApi {
	ta := &LogicApi{}
	ta.matchSessionMap = make(map[string]*innerMatchPlayer)
	ta.chanMatchPlay = make(map[string]chan *innerMatchPlayer)
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
	sid       string
	lv        int
	session   *hNet.Session
	other     *innerMatchPlayer
	isSuccess chan bool
}

/**
匹配算法
返回true 则为 这俩玩家可以被匹配
返回false 则为 这俩玩家不可以被匹配
*/
func (this *LogicApi) MatchRule(p1 *innerMatchPlayer, p2 *innerMatchPlayer) bool {
	if p1.sid == p2.sid {
		return false
	}

	if p1.lv-p2.lv >= -3 || p1.lv-p2.lv <= 3 {
		return true
	}
	return false
}

func (this *LogicApi) MatchSuccess(p1 *innerMatchPlayer, p2 *innerMatchPlayer) {
	this.rwLock.Lock()
	p1.other = p2
	p2.other = p1
	delete(this.matchSessionMap, p1.sid)
	delete(this.matchSessionMap, p2.sid)
	this.rwLock.Unlock()
	p2.isSuccess <- true
}

func (this *LogicApi) MatchTimer(curMatchPlayer *innerMatchPlayer, chanOther chan<- *innerMatchPlayer) {
	timer := time.NewTicker(time.Second)
	var other *innerMatchPlayer = nil
	passTime := 0
	// 每秒轮训查找一次匹配列表中
	for {
		select {
		case <-timer.C:
			{
				if passTime == 5 {
					timer.Stop()
					this.rwLock.Lock()
					delete(this.matchSessionMap, curMatchPlayer.sid)
					this.rwLock.Unlock()
					other = &innerMatchPlayer{
						sid:     "",
						lv:      99,
						session: nil,
					}
					hLog.Debug("匹配超时")
					goto Loop
				}

				this.rwLock.RLock()

				if _, ok := this.matchSessionMap[curMatchPlayer.sid]; !ok {
					timer.Stop()
					this.rwLock.RUnlock()
					other = curMatchPlayer.other
					goto Loop
				}

				for _, item := range this.matchSessionMap {
					if this.MatchRule(curMatchPlayer, item) {
						this.rwLock.RUnlock()
						this.MatchSuccess(curMatchPlayer, item)
						timer.Stop()
						other = item
						hLog.Debug("匹配真人")
						goto Loop
					}
				}
				this.rwLock.RUnlock()
				passTime++
			}
		case <-curMatchPlayer.isSuccess:
			hLog.Debug("自己被别人匹配了")
			// 自己被别人匹配了
			timer.Stop()
			other = curMatchPlayer.other
			goto Loop
		}
	}
Loop:
	hLog.Debug("OOOOOOO 哈哈哈哈哈哈哈哈")
	chanOther <- other
}

func (this *LogicApi) Match(session *hNet.Session, message *MatchMessage) {
	fmt.Println("来消息了  匹配")
	r := &MatchResMessage{
		CommonResMessage{
			Statue: CODE_OK,
			Msg:    "",
		}, components.MatchPlayInfo{
			NickName: "",
			HeadUrl:  "",
			Lv:       0,
		},
	}

	session.SetProperty("lv", 1)

	errReply := func(msg string) {
		r.Statue = CODE_ERROR
		r.Msg = msg
		this.Reply(session, r)
	}

	lv, ok := session.GetProperty("lv")
	if !ok {
		hLog.Error("网关服务器 匹配 获取玩家等级异常")
		errReply("网关服务器 匹配 获取玩家等级异常")
		return
	}
	this.rwLock.Lock()
	if _, ok := this.matchSessionMap[session.Id]; !ok {
		this.matchSessionMap[session.Id] = &innerMatchPlayer{
			sid:       session.Id,
			lv:        lv.(int),
			session:   session,
			isSuccess: make(chan bool),
			other:     nil,
		}
		this.chanMatchPlay[session.Id] = make(chan *innerMatchPlayer)
	} else {
		errReply("匹配中")
		return
	}
	this.rwLock.Unlock()
	go this.MatchTimer(this.matchSessionMap[session.Id], this.chanMatchPlay[session.Id])
	otherPlayer := <-this.chanMatchPlay[session.Id]

	/**
	匹配完成
	1.组装玩家1的数据
	2.组装玩家2的数据
	3.判断连接是否正常
	4.一处通道, 和 匹配玩家信息的map
	5.返回玩家
	*/
	r.NickName = session.Id
	r.Lv = otherPlayer.lv
	r.HeadUrl = "https://www.baidu.com"
	this.rwLock.Lock()
	close(this.chanMatchPlay[session.Id])
	delete(this.chanMatchPlay, session.Id)
	this.rwLock.Unlock()
	fmt.Println("发送数据", session.IsClose())
	if !session.IsClose() {
		this.Reply(session, r)
	}
}
