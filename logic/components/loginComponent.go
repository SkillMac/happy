package components

import (
	"../../hActor"
	"../../hBaseComponent"
	"../../hECS"
	"../../hLog"
	"time"

	//"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserInfo struct {
	ID         bson.ObjectId `bson:"_id"`
	OpenId     string        `bson:"openid"`
	NickName   string        `bson:"username"`
	HeadUrl    string        `bson:"headurl"`
	Lv         int           `bson:"lv"`
	Score      int           `bson:"score"`
	CreateTime int64         `bson:"createtime"`
	LastSign	int64
}

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

//func (this *LoginComponent) Start(ctx *hEcs.Context) {
/*
	这个是插入的逻辑
	err := hBaseComponent.Modle.M.Insert("users", &UserInfo{
		ID:       bson.NewObjectId(),
		NickName: "小熊",
		Lv:       1,
	})
	if err != nil {
		hLog.Info("插入用户失败")
		//message.Reply("插入用户失败")
	}
*/

/*
	查找返回的结果
		map[_id:ObjectIdHex("5d6f8dcde9c4c182be64bad9") lv:1 username:小熊]
	方式一
	hBaseComponent.Modle.M.DBFindOne("users", bson.M{"username": "小熊"}, func(a bson.M) error {
		if a != nil {
			hLog.Info(a)
			return nil
		} else {
			hLog.Error("FineOne err")
			return errors.New("FineOne err")
		}
	})

	方式二:
	a := &UserInfo{}
	hBaseComponent.Modle.M.FindOne("users", bson.M{"username": "小熊"}, a)
	hLog.Info("xxxxxxxx", a)

*/

//}

func (this *LoginComponent) Login(message *hActor.ActorMessageInfo) error {
	userInfo := &UserInfo{}
	hLog.Info("来啦,老弟=====>", message.Message.Data[1])
	err := hBaseComponent.Modle.M.FindOne("users", bson.M{"username": message.Message.Data[1]}, userInfo)
	if err != nil {
		message.Reply("没有该用户,创建新用户")
		err := hBaseComponent.Modle.M.Insert("users", &UserInfo{
			ID:         bson.NewObjectId(),
			OpenId:     message.Message.Data[0].(string),
			NickName:   message.Message.Data[1].(string),
			HeadUrl:    message.Message.Data[2].(string),
			Lv:         1,
			Score:      0,
			CreateTime: time.Now().Unix(),
		})
		if err != nil {
			return message.Reply("插入用户失败")
		} else {
			return message.Reply("插入用户成功")
		}
	} else {
		return message.Reply("查找用户成功")
	}
}
