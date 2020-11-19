package grawt

import (
	"sync"
)

type CloseHandler struct {
	waiter      *Waiter
	Quit        chan struct{}
	active      bool
	autoDone    bool
	wgDone      bool
	handlerFunc *func()
	mu          *sync.Mutex
}

func (ch *CloseHandler) Halt(err error) {
	ch.waiter.Halt(err)
}

func (ch *CloseHandler) Done() {
	ch.waiter.terminateHandler(ch, true)
}
