package hTimer

import (
	"sync"
	"time"
)

/**
时间轮
*/
type TimeWheel struct {
	lock      sync.Mutex
	t         time.Duration
	maxT      time.Duration
	ticker    *time.Ticker
	timeWheel []chan struct{}
	currPos   int
}

var (
	timerMap map[time.Duration]*TimeWheel
	mapLock  = &sync.Mutex{}
	accuracy = 20 // means max 1/20 deviation
)

/**
like time.After
*/
func After(t time.Duration) <-chan struct{} {
	mapLock.Lock()
	defer mapLock.Unlock()
	if v, ok := timerMap[t]; ok {
		return v.After(t)
	}
	v := NewTimeWheel(t/time.Duration(accuracy), accuracy+1)
	timerMap[t] = v
	return v.After(t)
}

func SetAccuracy(a int) {
	accuracy = a
}

func (this *TimeWheel) After(timeout time.Duration) <-chan struct{} {
	if timeout >= this.maxT {
		panic("timeout is bigger than maxT")
	}
	pos := int(timeout / this.t)
	if pos > 0 {
		pos--
	}
	this.lock.Lock()
	pos = (this.currPos + pos) % len(this.timeWheel)
	c := this.timeWheel[pos]
	this.lock.Unlock()
	return c
}

func (this *TimeWheel) run() {
	for range this.ticker.C {
		this.lock.Lock()
		oldestC := this.timeWheel[this.currPos]
		this.timeWheel[this.currPos] = make(chan struct{})
		this.currPos = (this.currPos + 1) % len(this.timeWheel)
		this.lock.Unlock()
		close(oldestC)
	}
}

func (this *TimeWheel) Stop() {
	this.ticker.Stop()
}

/**
New
*/
func NewTimeWheel(t time.Duration, size int) *TimeWheel {
	tw := &TimeWheel{t: t, maxT: t * time.Duration(size)}
	tw.timeWheel = make([]chan struct{}, size)
	for i := range tw.timeWheel {
		tw.timeWheel[i] = make(chan struct{})
	}
	tw.ticker = time.NewTicker(t)
	go tw.run()
	return tw
}

func init() {
	timerMap = make(map[time.Duration]*TimeWheel)
}
