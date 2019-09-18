package hCluster

import (
	"custom/happy/hCommon"
	"errors"
	"sync"
)

const (
	SELECTOR_TYPE_GROUP    SelectorType = "Group"
	SELECTOR_TYPE_DEFAULT  SelectorType = "Default"
	SELECTOR_TYPE_MIN_LOAD SelectorType = "MinLoad"
	SELECTOR_TYPE_CUSTOM   SelectorType = "Custom"
	SELECTOR_TYPE_RANDOM   SelectorType = "Random"
)

type SelectorType = string

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

type Selector map[string]*NodeInfo

var ErrNoAvailableNode = errors.New("query string wrong")

// 0 选择模式 1 AppName 2 role
func (this Selector) DoQuery(query []string, detail bool, locker *sync.RWMutex, selector ...func(SourceGroup) int) ([]*InquiryReply, error) {
	length := len(query)
	if length != 3 || query[0] == "" {
		return nil, ErrNoAvailableNode
	}

	err := errors.New("no available node ")
	var reply = make([]*InquiryReply, 0)
	locker.RLock()
	for nodeName, nodeInfo := range this {
		if nodeInfo.AppName == query[1] {
			for _, role := range nodeInfo.Role {
				if role == query[2] {
					if detail {
						reply = append(reply, &InquiryReply{Node: nodeName, Info: nodeInfo.Info})
					} else {
						reply = append(reply, &InquiryReply{Node: nodeName})
					}
					err = nil
					break
				}
			}
			if err == nil && query[0] != SELECTOR_TYPE_GROUP {
				break
			}
		}
	}
	locker.RUnlock()

	switch query[0] {
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
	default:

	}
	return reply, err
}
