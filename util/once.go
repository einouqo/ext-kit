package util

import (
	"sync"
	"sync/atomic"
)

type Once struct {
	mu   sync.Mutex
	done atomic.Value
	f    func()
}

func NewOnce(f func()) *Once {
	return &Once{f: f}
}

func (o *Once) Exec() {
	if o.done.Load() != nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.done.Load() != nil {
		return
	}
	o.f()
	o.done.Store(true)
}
