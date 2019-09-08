package api

import (
	"../../hActor"
	"../../hCluster"
	"../../hLog"
	"../../hNet"
	"../../hNet/messageProtocol"
	"../components"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
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

//微信登陆的相关逻辑
type ReqWxObj struct {
	Openid    string `json:"openid"`
	SessonKey string `json:"sesson_key"`
	Unionid   string `json:"unionid"`
	Errcode   int    `json:"errcode"`
	ErrMsg    string `json:"err_msg"`
}

func loginWX(code string) (wxInfo ReqWxObj, err error) {
	appId := "wx5cd18aeada5cf68b"
	appSecret := "03fb90bc2d468e4dc7da0d6553a993f3"
	url := "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
	resp, err := http.Get(fmt.Sprintf(url, appId, appSecret, code))
	if err != nil {
		return wxInfo, err
	}
	defer resp.Body.Close()
	err = BindJson(resp.Body, &wxInfo)
	if err != nil {
		return wxInfo, err
	}
	if wxInfo.Errcode != 0 {
		return wxInfo, errors.New(fmt.Sprintf("code:%d,errmsg:%s", wxInfo.Errcode, wxInfo.ErrMsg))
	}
	return wxInfo, nil
}

func BindJson(body io.ReadCloser, retMap interface{}) error {
	bodybuf, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	json.Unmarshal(bodybuf, &retMap)
	return nil
}

func (this *LogicApi) Login(sess *hNet.Session, message *LoginMessage) {
	hLog.Info("message  ", message)
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

	wxInfo := ""
	if message.JsCode != "123" {
		wxInfo, err := loginWX(message.JsCode)
		hLog.Info("wxInfo,err=====>", wxInfo, err)
	} else {
		wxInfo = "openid111111111111111"
	}

	reply, err := serviceCaller.Call("login", components.Service_Login_Login, message.NickName, message.HeadUrl, message.JsCode, wxInfo)
	if err != nil {
		hLog.Debug(err)
		errReply("登录失败服务器登录节40013点异常")
		return
	} else {
		sess.SetProperty("userInfo", message)
	}
	hLog.Info("reply=====>", reply)
	this.Reply(sess, &LoginResMessage{
		Statue: CODE_OK,
		Msg:    reply[0].(string),
	})
}

/**
匹配相关的逻辑
*/

type innerMatchPlayer struct {
	sid         string            // 回话ID
	lv          int               // 玩家的等级
	session     *hNet.Session     // 存储回话引用
	other       *innerMatchPlayer // 要对战的玩家
	isSuccess   chan bool         // 是否被别人匹配 的信号
	isShootBall bool              // 是够首先发球
	isMatched   bool              // 是否别人匹配
	roomId      int               // 房间ID
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

func (this *LogicApi) MatchSuccess(p1 *innerMatchPlayer, p2 *innerMatchPlayer, caller *hActor.ActorServiceCaller) {
	this.rwLock.Lock()
	p1.other = p2
	p2.other = p1
	delete(this.matchSessionMap, p1.sid)
	delete(this.matchSessionMap, p2.sid)
	this.rwLock.Unlock()

	r := &MatchResMessage{
		CommonResMessage{
			Statue: CODE_OK,
			Msg:    "",
		}, components.MatchPlayInfo{
			NickName:    "",
			HeadUrl:     "",
			Lv:          0,
			IsShootBall: false,
			RoomId:      -1,
			CrystalInfo: "",
		},
	}
	reply, err := caller.Call("room", components.Service_RoomManager_NewRoom, p1.sid, p2.sid)
	if err != nil {
		r.Statue = CODE_ERROR
		r.Msg = "Match 匹配服务器呼叫创建房间失败"
		this.Reply(p1.session, r)
		reply[0] = -1
	}

	roomId := reply[0].(int)
	p1.roomId = roomId
	p2.roomId = roomId

	flagP := this.generateRangeNum(1, 100) > 50
	p1.isShootBall = flagP
	p2.isShootBall = !flagP
	p2.isSuccess <- true
}

func (this *LogicApi) generateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

func (this *LogicApi) MatchTimer(curMatchPlayer *innerMatchPlayer, chanOther chan<- *innerMatchPlayer, caller *hActor.ActorServiceCaller) {
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
						sid:         "机器人",
						lv:          99,
						session:     nil,
						other:       curMatchPlayer,
						isShootBall: false,
						isMatched:   false,
						roomId:      -1,
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
						this.MatchSuccess(curMatchPlayer, item, caller)
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
			curMatchPlayer.isMatched = true
			goto Loop
		}
	}
Loop:
	hLog.Info("OOOOOOO 哈哈哈哈哈哈哈哈")
	chanOther <- other
}

func (this *LogicApi) Match(session *hNet.Session, message *MatchMessage) {
	hLog.Info("来消息了  匹配")
	r := &MatchResMessage{
		CommonResMessage{
			Statue: CODE_OK,
			Msg:    "",
		}, components.MatchPlayInfo{
			NickName:    "",
			HeadUrl:     "",
			Lv:          0,
			IsShootBall: false,
			RoomId:      -1,
			CrystalInfo: "",
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
			sid:         session.Id,
			lv:          lv.(int),
			session:     session,
			isSuccess:   make(chan bool),
			other:       nil,
			isShootBall: false,
			isMatched:   false,
			roomId:      -1,
		}
		this.chanMatchPlay[session.Id] = make(chan *innerMatchPlayer)
	} else {
		errReply("匹配中")
		return
	}
	this.rwLock.Unlock()

	serviceCaller, err := this.Upgrade(session)
	if err != nil {
		errReply("服务器Match 回话转换失败")
		return
	}

	go this.MatchTimer(this.matchSessionMap[session.Id], this.chanMatchPlay[session.Id], serviceCaller)
	otherPlayer := <-this.chanMatchPlay[session.Id]

	if otherPlayer.roomId == -1 && otherPlayer.sid != "机器人" {
		errReply("Match 匹配服务器 房间创建失败")
		return
	}

	/**
	匹配完成
	1.组装玩家1的数据
	2.组装玩家2的数据
	3.判断连接是否正常
	4.一处通道, 和 匹配玩家信息的map
	5.返回玩家
	*/
	session.SetProperty("OtherPlayer", otherPlayer.session)
	r.IsShootBall = otherPlayer.isShootBall
	r.NickName = otherPlayer.sid
	r.Lv = otherPlayer.lv
	r.HeadUrl = "https://wx.qlogo.cn/mmopen/vi_32/mtFonGxkxLZwLC31ibZJJuMWicfy4XGhazGBEic2Db8OH2JEmosEDRvyq0EEOx5uKqT1eTU8uk3qYRpvkzlsTA5Ig/132"
	this.rwLock.Lock()
	close(this.chanMatchPlay[session.Id])
	delete(this.chanMatchPlay, session.Id)
	this.rwLock.Unlock()
	r.CrystalInfo =r.MatchCrystalInfo()
	//fmt.Println("MatchCrystalInfo 11111111111111111111111", r.MatchCrystalInfo())
	r.RoomId = otherPlayer.roomId

	fmt.Println("发送数据", session.IsClose())
	this.Reply(session, r)
}

/**
创建房间的逻辑
*/

func (this *LogicApi) CreateRoom(session *hNet.Session, message *CreateRoomMessage) {
	r := &CreateRoomResMessage{
		CommonResMessage{
			Statue: CODE_OK,
			Msg:    "",
		},
		-1,
		"",
	}

	errReply := func(msg string) {
		r.Statue = CODE_ERROR
		r.Msg = msg
		this.Reply(session, r)
	}

	serviceCaller, err := this.Upgrade(session)
	if err != nil {
		errReply("服务器CreateRoom 回话转换失败")
		return
	}

	reply, err := serviceCaller.Call("room", components.Service_RoomManager_NewRoom, session.Id, "")
	if err != nil {
		hLog.Error(err)
		errReply("匹配 呼叫 房间服务器创建房间的函数失败")
		return
	}

	r.RoomId = reply[0].(int)
	r.CrystalInfo = reply[1].(interface{})
	this.Reply(session, r)
}

func (this *LogicApi) JoinRoom(session *hNet.Session, message *JoinRoomMessage) {
	r := &JoinRoomResMessage{
		CommonResMessage{
			Statue: CODE_OK,
			Msg:    "",
		},
	}

	errReply := func(msg string) {
		r.Statue = CODE_ERROR
		r.Msg = msg
		hLog.Info(msg)
		this.Reply(session, r)
	}

	serviceCaller, err := this.Upgrade(session)
	if err != nil {
		errReply("服务器 JoinRoom 回话转换失败")
		return
	}

	roomId := message.RoomId

	reply, err := serviceCaller.Call("room", components.Service_RoomManager_JoinRoom, roomId, session.Id)
	if err != nil {
		hLog.Info(err)
		errReply("匹配 呼叫 房间服务器加入房间的函数失败")
		return
	}
	r.Msg = reply[1].(string)

	this.Reply(session, r)
}

func (this *LogicApi) deleteRoom(session *hNet.Session, message *DeleteRoomMessage) {
	r := &DeleteRoomResMessage{
		CommonResMessage{
			Statue: CODE_OK,
			Msg:    "",
		},
	}
	errReply := func(msg string) {
		r.Statue = CODE_ERROR
		r.Msg = msg
		hLog.Info(msg)
		this.Reply(session, r)
	}
	serviceCaller, err := this.Upgrade(session)
	if err != nil {
		errReply("服务器 JoinRoom 回话转换失败")
		return
	}

	roomId := message.RoomId

	reply, err := serviceCaller.Call("room", components.Service_RoomManager_deleteRoom, roomId, session.Id)
	if err != nil {
		hLog.Info(err)
		errReply("匹配 呼叫 房间服务器删除房间的函数失败")
		return
	}
	r.Msg = reply[1].(string)
	this.Reply(session, r)

}

/**errReply
同步逻辑
*/

func (this *LogicApi) getServiceCall(session *hNet.Session, errReply func(msg string), MethodName string) *hActor.ActorServiceCaller {
	serviceCaller, err := this.Upgrade(session)
	if err != nil {
		errReply("服务器 " + MethodName + "回话转换失败")
		return serviceCaller
	}
	return nil
}

func (this *LogicApi) SyncData(session *hNet.Session, message *SyncMessage) {
	r := &SyncResMessage{
		CommonResMessage{
			Statue: CODE_OK,
			Msg:    "",
		}, *message,
	}

	errReply := func(msg string) {
		r.Statue = CODE_ERROR
		r.Msg = msg
		hLog.Info(msg)
		this.Reply(session, r)
	}

	//serviceCaller := this.getServiceCall(session, errReply, "SyncData")

	/**
	注意判断要同步玩家是否掉线
	*/
	otherS, ok := session.GetProperty("OtherPlayer")
	if !ok {
		//hLog.Info("同步数据 敌方玩家 回话获取失败")
		errReply("同步数据 敌方玩家 回话获取失白")
		return
	}

	otherSession := otherS.(*hNet.Session)

	if !otherSession.IsClose() {
		this.Reply(otherSession, r)
	} else {
		errReply("对方掉线")
	}
}
