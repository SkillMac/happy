package api

import (
	"../../hCluster"
	"../../hNet"
	"../../hNet/messageProtocol"
	"fmt"
	"reflect"
)

type TestMessage struct {
	Name string
}

var TestId2mt = map[reflect.Type]uint32{
	reflect.TypeOf(&TestMessage{}): 1,
}

type TestApi struct {
	hNet.ApiBase
	nodeComponent hCluster.NodeComponent
}

func NewTestApi() *TestApi {
	ta := &TestApi{}
	ta.Instance(ta).SetMT2ID(TestId2mt).SetProtocol(&messageProtocol.JsonProtocol{})
	return ta
}

func (this *TestApi) Hello(sess *hNet.Session, message *TestMessage) {
	println(fmt.Sprintf("hello %s", message.Name))
	sess.Emit(1, []byte(fmt.Sprintf("hello client %s", message.Name)))
}
