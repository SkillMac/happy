package hCluster

import (
	"custom/happy/hCommon"
	"errors"
	"fmt"
	"github.com/lafikl/liblb/bounded"
	"github.com/lafikl/liblb/p2c"
	"github.com/lafikl/liblb/r2"
	"sync"
)

const (
	SELECTOR_TYPE_GROUP             SelectorType = "Group"
	SELECTOR_TYPE_DEFAULT           SelectorType = "Default"
	SELECTOR_TYPE_MIN_LOAD          SelectorType = "MinLoad"
	SELECTOR_TYPE_CUSTOM            SelectorType = "Custom"
	SELECTOR_TYPE_RANDOM            SelectorType = "Random"
	SELECTOR_TYPE_POLLING           SelectorType = "Polling"
	SELECTOR_TYPE_P2C               SelectorType = "P2C"
	SELECTOR_TYPE_CONST_HASHING_BL  SelectorType = "ConstHashingBoundLoad"
	SELECTOR_TYPE_CONST_HASHING_BLW SelectorType = "ConstHashingBoundLoadWithWeight"
)

type SelectorType = string

var r2LB = r2.New()
var p2cLB = p2c.New()
var boundedLB = bounded.New()

type SourceGroup []*InquiryReply

//最小负载：cpu * 80% + mem * 20%
func (this SourceGroup) SelectMinLoad() int {
	var min float64 = 1
	var index int = -1
	for i, info := range this {
		var cpu, mem float64 = 1, 1
		if v, ok := info.Info["cpu"]; ok {
			cpu = v
		}
		if v, ok := info.Info["mem"]; ok {
			mem = v
		}
		sum := cpu*0.8 + mem*0.2
		if sum <= min {
			min = sum
			index = i
		}
	}
	return index
}

func (this SourceGroup) Random() int {
	d := hCommon.GenRandom(0, len(this), 1)
	if d == nil {
		d = []int{-1}
	}
	return d[0]
}

func (this SourceGroup) Polling() int {
	addr, err := r2LB.Balance()

	if err != nil {
		return -1
	}

	for i, info := range this {
		if addr == info.Node {
			return i
		}
	}

	return -1
}

func (this SourceGroup) P2C(key string) int {
	defer func() {
		p2cLB.Done(key)
	}()
	host, err := p2cLB.Balance(key)
	if err != nil {
		return -1
	}

	for i, info := range this {
		if host == info.Node {
			return i
		}
	}

	return -1
}

func (this SourceGroup) CHashingBL(key string) int {
	defer func() {
		boundedLB.Done(key)
	}()
	host, err := boundedLB.Balance(key)
	if err != nil {
		return -1
	}

	fmt.Println("lllllllllllllllllllllll", host)

	for i, info := range this {
		if host == info.Node {
			return i
		}
	}

	return -1
}

func (this SourceGroup) CHashingBLW(key string) int {
	return 0
}

func (this SourceGroup) DoQuery(selectorType SelectorType) int {
	switch selectorType {
	case SELECTOR_TYPE_DEFAULT, SELECTOR_TYPE_RANDOM:
		return this.Random()
	case SELECTOR_TYPE_MIN_LOAD:
		return this.SelectMinLoad()
	case SELECTOR_TYPE_CUSTOM:
		if len(this) == 0 {
			return -1
		}
		return 0
	default:
		return -1
	}
}

type Selector map[string]*NodeInfo

var ErrNoAvailableNode = errors.New("query string wrong")

// 0 AppName 1 role 2 选择模式
// 3 如果是 P2C 或者是 哈希一致性算法 是需要一个 唯一 key
func (this Selector) DoQuery(query []string, detail bool, locker *sync.RWMutex, selector ...func(SourceGroup) int) ([]*InquiryReply, error) {
	length := len(query)
	if length < 3 || query[2] == "" {
		return nil, ErrNoAvailableNode
	}
	err := errors.New("no available node ")
	var reply = make([]*InquiryReply, 0)
	locker.RLock()
	var hosts []string
	for nodeName, nodeInfo := range this {
		if nodeInfo.AppName == query[0] {
			for _, role := range nodeInfo.Role {
				if role == query[1] {
					if detail {
						reply = append(reply, &InquiryReply{Node: nodeName, Info: nodeInfo.Info, CustomData: nodeInfo.CustomData})
					} else {
						reply = append(reply, &InquiryReply{Node: nodeName, CustomData: nodeInfo.CustomData})
					}
					hosts = append(hosts, nodeName)
					err = nil
					break
				}
			}
			//if err == nil && query[0] != SELECTOR_TYPE_GROUP {
			//	break
			//}
		}
	}
	locker.RUnlock()

	switch query[2] {
	case SELECTOR_TYPE_MIN_LOAD:
		var index = -1
		index = SourceGroup(reply).SelectMinLoad()
		if index != -1 {
			reply = []*InquiryReply{reply[index]}
		}
	case SELECTOR_TYPE_GROUP:
	case SELECTOR_TYPE_CUSTOM:
		var index = -1
		if len(selector) == 0 {
			err = errors.New("custom selector is empty")
		}
		index = selector[0](SourceGroup(reply))
		reply = []*InquiryReply{reply[index]}
	case SELECTOR_TYPE_DEFAULT, SELECTOR_TYPE_RANDOM:
		var index = -1
		index = SourceGroup(reply).Random()
		if index == -1 {
			err = errors.New("[Select-DoQuery] Random index = -1")
		} else if index != -1 {
			reply = []*InquiryReply{reply[index]}
		}
	case SELECTOR_TYPE_POLLING:
		var index = SourceGroup(reply).Polling()
		if index == -1 {
			err = errors.New("[Select-DoQuery] Polling index = -1")
		} else if index != -1 {
			reply = []*InquiryReply{reply[index]}
		}
	case SELECTOR_TYPE_P2C:
		var index = SourceGroup(reply).P2C(query[3])
		if index == -1 {
			err = errors.New("[Select-DoQuery] P2C index = -1")
		} else if index != -1 {
			reply = []*InquiryReply{reply[index]}
		}
	case SELECTOR_TYPE_CONST_HASHING_BL:
		var index = SourceGroup(reply).CHashingBL(query[3])
		if index == -1 {
			err = errors.New("[Select-DoQuery] CHashingBL index = -1")
		} else if index != -1 {
			reply = []*InquiryReply{reply[index]}
		}
	case SELECTOR_TYPE_CONST_HASHING_BLW:
		var index = SourceGroup(reply).CHashingBLW(query[3])
		if index == -1 {
			err = errors.New("[Select-DoQuery] CHashingBLW index = -1")
		} else if index != -1 {
			reply = []*InquiryReply{reply[index]}
		}
	default:
	}
	return reply, err
}
