package components

import (
	"../../hActor"
	"../../hECS"
	"math/rand"
	"sync"
	"time"
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
func RandNum(num int) int {

	if r := rand.Intn(num); r == 0 {
		return r + 1
	} else {
		return r
	}
}

func RandNumScope(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}