package hActor

import (
	"custom/happy/hCluster"
	"custom/happy/hNet"
	"sync"
)

const (
	LOCAL_SERVICE = ""
)

type ActorServiceCaller struct {
	locker   sync.RWMutex
	proxy    *ActorProxyComponent
	services map[string]*ActorService
}

func NewActorServiceCaller(proxy *ActorProxyComponent) *ActorServiceCaller {
	return &ActorServiceCaller{proxy: proxy, services: make(map[string]*ActorService)}
}

func NewActorServiceCallerFromSession(sess *hNet.Session, proxy *ActorProxyComponent) *ActorServiceCaller {
	g, ok := sess.GetProperty("ActorServiceCaller")
	if ok {
		return g.(*ActorServiceCaller)
	}
	sc := NewActorServiceCaller(proxy)
	sess.SetProperty("ActorServiceCaller", sc)
	return sc
}

func (this *ActorServiceCaller) Call(role string, serviceName string, args ...interface{}) ([]interface{}, error) {
	var err error
	//优先尝试缓存客户端，避免反复查询，尽量去中心化
	service, ok := this.services[serviceName]
	if ok {
		res, err := service.Call(args...)
		if err != nil {
			delete(this.services, serviceName)
		} else {
			return res, err
		}
	}
	//无缓存，或者通过缓存调用失败，重新查询调用
	service, err = this.proxy.GetActorService(role, serviceName)
	if err != nil {
		return nil, err
	}
	this.services[serviceName] = service
	res, err := service.Call(args...)
	if err != nil {
		delete(this.services, serviceName)
	}
	return res, err
}

func (this *ActorServiceCaller) CallWait(role string, serviceName string, args ...interface{}) ([]interface{}, error) {
	var err error
	//优先尝试缓存客户端，避免反复查询，尽量去中心化
	service, ok := this.services[serviceName]
	if ok {
		res, err := service.CallWait(args...)
		if err != nil {
			delete(this.services, serviceName)
		} else {
			return res, err
		}
	}
	//无缓存，或者通过缓存调用失败，重新查询调用
	service, err = this.proxy.GetActorService(role, serviceName)
	if err != nil {
		return nil, err
	}
	this.services[serviceName] = service
	res, err := service.CallWait(args...)
	if err != nil {
		delete(this.services, serviceName)
	}
	return res, err
}

/*
* 用于其他 需要传参的负载均衡算法
* params {host client, selectType}
 */
func (this *ActorServiceCaller) CallWithSelectType(role string, serviceName string, selectType []hCluster.SelectorType, args ...interface{}) ([]interface{}, error) {
	var err error
	//优先尝试缓存客户端，避免反复查询，尽量去中心化
	service, ok := this.services[serviceName]
	if ok {
		res, err := service.Call(args...)
		if err != nil {
			delete(this.services, serviceName)
		} else {
			return res, err
		}
	}
	//无缓存，或者通过缓存调用失败，重新查询调用
	service, err = this.proxy.GetRemoteActorService(role, serviceName, selectType...)
	if err != nil {
		return nil, err
	}
	this.services[serviceName] = service
	res, err := service.Call(args...)
	if err != nil {
		delete(this.services, serviceName)
	}
	return res, err
}
