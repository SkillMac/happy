package hNet

import (
	"context"
	"custom/happy/hECS"
)

type IHander interface {
	Listen() error
	Handle() error
	CheckClose()
	Destroy() error
}

type IPackageProtocol interface {
	Encode([]byte) (int, int)
	Decode(context.Context, []byte) ([]uint32, []byte)
}

type ILogicAPI interface {
	Init(parent ...*hEcs.Object)
	Route(session *Session, messageID uint32, data []byte)
	Reply(session *Session, message interface{})
	OnConnect(session *Session)
	OnDisconnect(session *Session)
}

type IWsConn interface {
	WriteMessage(messageType uint32, data []byte) error
	Addr() string
	Close() error
}
