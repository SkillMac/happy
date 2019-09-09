package api

import "reflect"
import "../components"

const (
	CODE_ERROR = -1
	CODE_OK    = 0
)

type CommonResMessage struct {
	Statue int
	Msg    string
}

type TestMessage struct {
	Name string
}

// 客户端向服务器发送的结构器
type LoginMessage struct {
	JsCode   string //jscode
	NickName string // 微信名字
	HeadUrl  string // 头像地址
}

// 登录成功后服务器响应的结构体
type LoginResMessage struct {
	Statue int
	Msg    string
}

// 匹配结构
type MatchMessage struct {
}

type MatchResMessage struct {
	CommonResMessage
	components.MatchPlayInfo
}

// 创建房间的结构体
type CreateRoomMessage struct {
}

type CreateRoomResMessage struct {
	CommonResMessage
	RoomId      int
	CrystalInfo interface{}
}

// 加入房间
type JoinRoomMessage struct {
	RoomId int
}

type JoinRoomResMessage struct {
	CommonResMessage
}

// 删除房间
type DeleteRoomMessage struct {
	RoomId int
}
type DeleteRoomResMessage struct {
	CommonResMessage
}

// 同步的函数
type vec2 struct {
	X float32
	Y float32
}

//type SyncCrystal struct {
//	Enemy  interface{}
//	Player interface{}
//	Broken interface{}
//}
type SyncMessage struct {
	Pos           vec2    // 玩家的位置
	Rotation      float32 // 玩家的旋转角度
	BallLineSpeed vec2    // 球的线速度
	BallPos       vec2    // 球的位置
	BallParent    string  // 球客户端的父节点
	Touch         bool    // 被发球方 是否触摸屏幕
	Broken        interface{}
	//CrystalInfo   SyncCrystal
}


type SyncResMessage struct {
	CommonResMessage
	SyncMessage
}

var Id2mt = map[reflect.Type]uint32{
	reflect.TypeOf(&TestMessage{}):          1,
	reflect.TypeOf(&CommonResMessage{}):     2,
	reflect.TypeOf(&LoginMessage{}):         3,
	reflect.TypeOf(&LoginResMessage{}):      4,
	reflect.TypeOf(&MatchMessage{}):         5,
	reflect.TypeOf(&MatchResMessage{}):      6,
	reflect.TypeOf(&CreateRoomMessage{}):    7,
	reflect.TypeOf(&CreateRoomResMessage{}): 8,
	reflect.TypeOf(&JoinRoomMessage{}):      9,
	reflect.TypeOf(&JoinRoomResMessage{}):   10,
	reflect.TypeOf(&SyncMessage{}):          11,
	reflect.TypeOf(&SyncResMessage{}):       12,
	reflect.TypeOf(&DeleteRoomMessage{}):    13,
	reflect.TypeOf(&DeleteRoomResMessage{}): 14,
}
