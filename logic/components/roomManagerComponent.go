package components

import (
	"../../hActor"
	"../../hECS"
)

type RoomManagerComponent struct {
	BaseComponent
	rooms    map[int]*RoomComponent
	roomRoot *hEcs.Object
	roomNum  int // 这一块后面会修改
}

func (this *RoomManagerComponent) Awake(ctx *hEcs.Context) {
	this.rooms = make(map[int]*RoomComponent)
	// 创建房间管理器根节点
	this.roomRoot = hEcs.NewObject("RoomManagerRoot")
	this.Parent().AddObject(this.roomRoot)
	this.AddHandler(Service_RoomManager_NewRoom, this.NewRoom, true)
}

func (this *RoomManagerComponent) getRoomNum() int {
	this.Locker.Lock()
	defer this.Locker.Unlock()
	this.roomNum++
	return this.roomNum
}

func (this *RoomManagerComponent) roomNumDecr() {
	this.Locker.Lock()
	defer this.Locker.Unlock()
	this.roomNum--
}

var Service_RoomManager_NewRoom = "NewRoom"

func (this *RoomManagerComponent) NewRoom(message *hActor.ActorMessageInfo) error {
	sid1 := message.Message.Data[0].(string)
	sid2 := message.Message.Data[0].(string)
	r := &RoomComponent{}
	id := this.getRoomNum()
	_, err := this.roomRoot.AddNewObjectWithComponent(r, string(id))
	if err != nil {
		return err
	}
	this.Locker.Lock()
	r.RoomID = id
	r.Sid = &[]string{sid1, sid2}
	this.rooms[id] = r
	this.Locker.Unlock()
	return message.Reply(r.RoomID)
}

var Service_RoomManager_JoinRoom = "JoinRoom"

func (this *RoomManagerComponent) CheckJoinRoom(roomId int, sid string) int {
	this.Locker.Lock()
	defer this.Locker.Unlock()
	roomComponent, ok := this.rooms[roomId]

	if !ok {
		return 1 // 房间不存在
	}

	if roomComponent.getSid()[1] != "" {
		return 2 // 房间满员
	}

	if roomComponent.getSid()[0] != "" {
		return 3 // 房主数据异常
	}

	return 0
}

func (this *RoomManagerComponent) JoinRoom(message *hActor.ActorMessageInfo) error {
	roomId := message.Message.Data[0].(int)
	sid := message.Message.Data[1].(string)

	code := this.CheckJoinRoom(roomId, sid)

	if code == 0 {
		return message.Reply(code, "加入成功")
	} else if code == 1 {
		return message.Reply(code, "房间不存在")
	} else if code == 2 {
		return message.Reply(code, "房间满员")
	} else if code == 3 {
		return message.Reply(code, "房主信息异常")
	}
	return nil
}
