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
	Nickname string // 微信名字
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

var Id2mt = map[reflect.Type]uint32{
	reflect.TypeOf(&TestMessage{}):      1,
	reflect.TypeOf(&CommonResMessage{}): 2,
	reflect.TypeOf(&LoginMessage{}):     3,
	reflect.TypeOf(&LoginResMessage{}):  4,
	reflect.TypeOf(&MatchMessage{}):     5,
	reflect.TypeOf(&MatchResMessage{}):  6,
}
