package hNet

import (
	"custom/happy/hCommon"
	"custom/happy/hECS"
	"custom/happy/hLog"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type MessageProtocol interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

type methodType struct {
	resv     reflect.Value
	method   reflect.Value
	argsType reflect.Type
}

var ErrNotInit = errors.New("this api is not initialized")
var ErrApiHandlerParamWrong = errors.New("this handler param wrong")
var ErrApiRepeated = errors.New("this ApiBase is  repeated")

// route mt2id 仅在初始化阶段写入，所以并发状态并无竞态
var route = map[uint32]*methodType{}
var mt2id = map[reflect.Type]uint32{}

type ApiBase struct {
	//注入子类
	instance interface{}
	protoc   MessageProtocol
	parent   *hEcs.Object
	isInit   bool
	SessMap  sync.Map
}

func (this *ApiBase) Instance(instance interface{}) *ApiBase {
	this.instance = instance
	return this
}

func (this *ApiBase) Init(parent ...*hEcs.Object) {
	if route == nil {
		route = map[uint32]*methodType{}
	}

	if len(parent) > 0 {
		this.parent = parent[0]
	}

	if this.protoc == nil || this.parent == nil || this.instance == nil {
		panic(ErrNotInit)
	}

	this.isInit = true

	this.RegisterGroup(this.instance)
}

func (this *ApiBase) checkInit() {
	if !this.isInit {
		panic(ErrNotInit)
	}
}

func (this *ApiBase) SetProtocol(protocol MessageProtocol) *ApiBase {
	this.protoc = protocol
	return this
}

func (this *ApiBase) GetProtocol() (protocol MessageProtocol) {
	return this.protoc
}

func (this *ApiBase) SetMT2ID(mtToId map[reflect.Type]uint32) *ApiBase {
	for key, value := range mtToId {
		if v, ok := mt2id[key]; ok {
			hLog.Error(fmt.Sprintf("this message [ %s ] id is repeated between [ %d ] and [ %d ]",
				key.Name(), v, value))
		} else {
			mt2id[key] = value
		}
	}
	return this
}

func (this *ApiBase) GetMT2ID() map[reflect.Type]uint32 {
	return mt2id
}

func (this *ApiBase) SetParent(parent *hEcs.Object) *ApiBase {
	this.parent = parent
	return this
}

func (this *ApiBase) GetParent() (*hEcs.Object, error) {
	var err error
	if this.parent == nil {
		err = errors.New("this api has not parent")
	}
	return this.parent, err
}

func (this *ApiBase) Route(sess *Session, messageID uint32, data []byte) {
	this.checkInit()
	defer hCommon.CheckError()
	if mt, ok := route[messageID]; ok {
		v := reflect.New(mt.argsType.Elem())
		err := this.protoc.Unmarshal(data, v.Interface())
		if err != nil {
			hLog.Debug(fmt.Sprintf("unmarshal message failed :%s ,%s", mt.argsType.Elem().Name(), err))
			return
		}
		var args []reflect.Value
		if mt.resv.IsNil() {
			args = []reflect.Value{
				mt.resv,
				reflect.ValueOf(sess),
				v,
			}
		} else {
			args = []reflect.Value{
				mt.resv,
				reflect.ValueOf(sess),
				v,
			}
		}
		mt.method.Call(args)
		return
	}
	hLog.Debug(fmt.Sprintf("this ApiBase:%d not found", messageID))
}

func (this *ApiBase) Reply(sess *Session, message interface{}) {
	this.checkInit()
	defer hCommon.CheckError()

	t := reflect.TypeOf(message)
	if id, ok := mt2id[t]; !ok {
		switch t.Kind() {
		case reflect.Struct:
			panic(errors.New(fmt.Sprintf("this message %s must be pointer,stead of &%s.", t.Name(), t.Name())))
		default:
			panic(errors.New(fmt.Sprintf("this message type: %s not be registered", t.Name())))
		}
	} else {
		m, err := this.protoc.Marshal(message)
		if err != nil {
			panic(err)
		}
		if !sess.IsClose() {
			panic(sess.Emit(id, m))
		}
	}
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var st = reflect.TypeOf(&Session{})

func (this *ApiBase) Register(handler interface{}) {
	this.checkInit()
	mValue := reflect.ValueOf(handler)
	mType := reflect.TypeOf(handler)
	paramsCount := mType.NumIn()
	if paramsCount != 2 {
		panic(ErrApiHandlerParamWrong)
	}
	sessType := mType.In(0)
	if sessType != st {
		return
	}

	argsType := mType.In(1)
	if !hCommon.IsExportedOrBuiltinType(argsType) {
		return
	}

	if index, ok := mt2id[argsType]; ok {
		if _, exist := route[index]; exist {
			panic(ErrApiRepeated)
		} else {
			route[index] = &methodType{
				method:   mValue,
				argsType: argsType,
			}
		}
	}
}

func (this *ApiBase) RegisterGroup(api interface{}) {
	this.checkInit()

	typ := reflect.TypeOf(api)

	//检查类型，如果是处理函数，改用 Register
	switch typ.Kind() {
	case reflect.Func:
		this.Register(api)
		return
	}

	hLog.Info(fmt.Sprintf("====== start to register API group: [ %s ] ======", typ.Elem().Name()))
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.

		if method.PkgPath != "" {
			continue
		}
		numin := mtype.NumIn()
		if numin != 3 {
			continue
		}

		sessType := mtype.In(1)
		if sessType != st {
			continue
		}

		argsType := mtype.In(2)
		if !hCommon.IsExportedOrBuiltinType(argsType) {
			continue
		}
		// Method needs one out.
		//if mtype.NumOut() != 1 {
		//	continue
		//}

		// The return type of the method must be error.
		//if returnType := mtype.Out(0); returnType != typeOfError {
		//	continue
		//}
		if index, ok := mt2id[argsType]; ok {
			if _, exist := route[index]; exist {
				panic(ErrApiRepeated)
			} else {
				route[index] = &methodType{
					resv:     reflect.ValueOf(api),
					method:   method.Func,
					argsType: argsType,
				}
			}
			hLog.Info(fmt.Sprintf("Add api: [ %s ], handler: [ %s.%s(*network.Session,*%s) ]", argsType.Elem().Name(), typ.Elem().Name(), mname, argsType.Elem().Name()))
		}
	}
	hLog.Info(fmt.Sprintf("======   register API group: [ %s ] end   ======", typ.Elem().Name()))
}

func (this *ApiBase) GetMessageType(message interface{}) (uint32, bool) {
	this.checkInit()
	id, ok := mt2id[reflect.TypeOf(message)]
	return id, ok
}

func (this *ApiBase) OnConnect(sess *Session) {
	this.SessMap.Store(sess.Id, &sess)
}

func (this *ApiBase) OnDisconnect(sess *Session) {
	this.SessMap.Delete(sess.Id)
}
