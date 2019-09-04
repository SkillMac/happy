package components

import (
	"../../hActor"
	"../../hECS"
	"fmt"
)

/*
type LoginMessage struct {
	nickname string // 微信名字
	headUrl string // 头像地址
}
**/

// 每个自定义组件里面必须报的三个属性
// 继承 ComponentBase 和 ActorBase(这个如果没有永奥可以不继承)
// 必须加读写锁
type LoginComponent struct {
	BaseComponent
}

var Service_Login_Login = "Login"

func (this *LoginComponent) Awake(ctx *hEcs.Context) {
	this.AddHandler(Service_Login_Login, this.Login, true)
}

func (this *LoginComponent) Login(message *hActor.ActorMessageInfo) error {
	fmt.Println("登录组件收到的数据", message.Message.Data[0])

	return message.Reply("我收到了消息")
}
