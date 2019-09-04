package components

import (
	"../../hActor"
	"../../hECS"
	"container/list"
)

type MatchComponent struct {
	BaseComponent
	waitList list.List
}

type innerMatchPlayer struct {
	sid string
	lv  int
	cb  func(cidArr []string) error
}

func (this *MatchComponent) Initialize() error {
	this.ActorInit(this.Parent())
	this.waitList = list.List{}
	return nil
}

func (this *MatchComponent) Awake(ctx *hEcs.Context) {
	this.AddHandler(Service_Match_Match, this.Match, true)
	//this.MatchTimer()
}

func (this *MatchComponent) MatchRule(p1 *innerMatchPlayer, p2 *innerMatchPlayer) bool {
	if p1.sid == p2.sid {
		return false
	}

	if p1.lv-p2.lv < 3 || p1.lv-p2.lv > -3 {
		return true
	}

	return false
}

func (this *MatchComponent) Do() {
	this.Locker.RLock()
	if this.waitList.Len() > 0 {
		p := this.waitList.Front()
		pVal := p.Value.(*innerMatchPlayer)

		for other := this.waitList.Front(); other != nil; other.Next() {
			otherVal := other.Value.(*innerMatchPlayer)
			if this.MatchRule(pVal, otherVal) {
				this.Locker.Lock()
				this.waitList.Remove(p)
				this.waitList.Remove(other)
				this.Locker.Unlock()
				sidArr := []string{pVal.sid, otherVal.sid}
				pVal.cb(sidArr)
				otherVal.cb(sidArr)
				return
			}
		}
		pVal.cb([]string{"", ""})
	}
	this.Locker.RUnlock()
}

//
//func (this *MatchComponent) MatchTimer() {
//	// 每一秒 检查一次
//	timer := time.NewTimer(time.Second)
//	for {
//		select {
//		case <-timer.C:
//			{
//				this.Do()
//			}
//		}
//	}
//}

var Service_Match_Match = "Match"

/*
type playInfo struct {
	NickName string
	HeadUrl  string
	Lv       int
}
*/
func (this *MatchComponent) Match(message *hActor.ActorMessageInfo) error {
	sid := message.Message.Data[0].(string)
	lv := message.Message.Data[1].(int)
	passTime := message.Message.Data[2].(int)
	if passTime == 0 {
		this.Locker.Lock()
		this.waitList.PushBack(&innerMatchPlayer{
			sid: sid,
			lv:  lv,
			cb: func(sidArr []string) error {
				return message.Reply(sidArr[0], sidArr[1])
			},
		})
		this.Locker.Unlock()
	} else if passTime == 10 {
		return message.Reply([]string{sid, ""})
	}
	this.Do()
	return nil
	//// 现在默认匹配机器人
	//return message.Reply(
	//	"机器人",
	//	"https://wx.qlogo.cn/mmopen/vi_32/mtFonGxkxLZwLC31ibZJJuMWicfy4XGhazGBEic2Db8OH2JEmosEDRvyq0EEOx5uKqT1eTU8uk3qYRpvkzlsTA5Ig/132",
	//	1,
	//)
}
