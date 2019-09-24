package hBaseComponent

import "sync"

type IBDestroy interface {
	CheckClose(group *sync.WaitGroup)
}
