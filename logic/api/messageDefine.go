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
	RoomId int
}

// 加入房间
type JoinRoomMessage struct {
	RoomId int
}

type JoinRoomResMessage struct {
	CommonResMessage
}

// 同步的函数
type vec2 struct {
	X float32
	Y float32
}
type SyncMessage struct {
	Pos           vec2
	Rotation      float32
	BallLineSpeed vec2
	BallPos       vec2
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
}
