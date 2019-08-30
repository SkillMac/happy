package hNet

import (
	"../hECS"
	"context"
)

type IHander interface {
	Listen() error
	Handle() error
}

type IPackageProtocol interface {
	Encode([]byte) (int, int)
	Decode(context.Context, []byte) ([]uint32, []byte)
}

type ILogicAPI interface {
	Init(parent ...*hEcs.Object)
	Route(session *Session, messageID uint32, data []byte)
	Reply(session *Session, message interface{})
}

type IWsConn interface {
	WriteMessage(messageType uint32, data []byte) error
	Addr() string
	Close() error
}
