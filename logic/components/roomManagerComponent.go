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
	r := &RoomComponent{}
	id := this.getRoomNum()
	_, err := this.roomRoot.AddNewObjectWithComponent(r, string(id))
	if err != nil {
		return err
	}
	this.Locker.Lock()
	r.RoomID = id
	this.rooms[id] = r
	this.Locker.Unlock()
	return message.Reply(r.RoomID)
}
