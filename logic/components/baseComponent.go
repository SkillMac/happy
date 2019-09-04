package components

import (
	"../../hActor"
	"../../hECS"
	"sync"
)

type BaseComponent struct {
	hEcs.ComponentBase
	hActor.ActorBase
	Locker sync.RWMutex
}

func (this *BaseComponent) Initialize() error {
	this.ActorInit(this.Parent())
	return nil
}
