package hNet

import (
	"errors"
	"sync"
)

type Session struct {
	rwLock         sync.RWMutex
	Id             string
	properties     map[string]interface{}
	conn           IWsConn
	postProcessing []func(sess *Session)
}

func (this *Session) AddPostProcessing(fn func(sess *Session)) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()

	this.postProcessing = append(this.postProcessing, fn)
}

func (this *Session) PostProcessing() {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()

	for _, fn := range this.postProcessing {
		fn(this)
	}
}

func (this *Session) RemoteAddr() string {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()

	return this.conn.Addr()
}

func (this *Session) Close() error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()

	return this.conn.Close()
}

func (this *Session) SetProperty(key string, value interface{}) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	this.properties[key] = value
}

func (this *Session) GetProperty(key string) (interface{}, bool) {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()

	p, ok := this.properties[key]
	return p, ok
}

func (this *Session) RemoveProperty(key string) {
	this.rwLock.Lock()
	defer this.rwLock.RUnlock()

	delete(this.properties, key)
}

var ErrSessionEissconnected = errors.New("this session is broken")

func (this *Session) Emit(messageType uint32, message []byte) error {
	if this.conn == nil {
		return ErrSessionEissconnected
	}
	return this.conn.WriteMessage(messageType, message)
}

func (this *Session) IsClose() bool {
	if this.conn == nil {
		return true
	}
	return false
}

/**
New
*/
func NewSeesion(id string, conn IWsConn) *Session {
	sess := &Session{
		rwLock:     sync.RWMutex{},
		Id:         id,
		properties: make(map[string]interface{}),
		conn:       conn,
	}
	return sess
}
