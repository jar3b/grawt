package grawt

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Waiter struct {
	sync.RWMutex
	blockingMode  bool
	waitGroup     sync.WaitGroup
	closeHandlers []*CloseHandler
	haltingFlag   bool
	haltingMutex  sync.RWMutex
}

func (w *Waiter) isHalting() bool {
	w.haltingMutex.RLock()
	defer w.haltingMutex.RUnlock()
	return w.haltingFlag
}

func (w *Waiter) setHalting(value bool) {
	w.haltingMutex.Lock()
	defer w.haltingMutex.Unlock()
	w.haltingFlag = value
}

func (w *Waiter) addHandler(f *func(), autoDone bool) *CloseHandler {
	ch := CloseHandler{
		waiter:      w,
		Quit:        make(chan struct{}, 1),
		active:      true,
		autoDone:    autoDone,
		wgDone:      false,
		handlerFunc: f,
		mu:          &sync.Mutex{},
	}

	w.Lock()
	w.waitGroup.Add(1)
	w.closeHandlers = append(w.closeHandlers, &ch)
	w.Unlock()

	return &ch
}

func (w *Waiter) terminateHandler(h *CloseHandler, forceWaitGroupDone bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.active {
		if !h.wgDone {
			w.waitGroup.Done()
		}
		return
	}

	if h.handlerFunc != nil && *h.handlerFunc != nil {
		(*h.handlerFunc)()
	}

	close(h.Quit)
	if h.autoDone || forceWaitGroupDone {
		w.waitGroup.Done()
		h.wgDone = true
	}

	h.active = false
}

func (w *Waiter) AddCloseHandler(f func(), waitForChannel bool) *CloseHandler {
	return w.addHandler(&f, !waitForChannel)
}

func (w *Waiter) Halt(err error) {
	if w.isHalting() {
		return
	}
	w.setHalting(true)

	w.RLock()
	for _, h := range w.closeHandlers {
		w.terminateHandler(h, false)
	}
	w.RUnlock()

	if err != nil {
		log.Errorf("Program was terminated with error: %v", err)
	} else {
		log.Info("Program was terminated gracefully.")
	}
	if !w.blockingMode {
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	w.setHalting(false)
}

func (w *Waiter) Wait() {
	w.blockingMode = true
	log.Info("Waiting...")
	w.waitGroup.Wait()
	log.Info("Terminated")
}

func (w *Waiter) onSignal(sig os.Signal) {
	log.Infof("Received signal '%s'! Exiting...", sig.String())
	w.Halt(nil)
}

func NewWaiter() *Waiter {
	w := Waiter{
		sync.RWMutex{},
		false,
		sync.WaitGroup{},
		make([]*CloseHandler, 0),
		false,
		sync.RWMutex{},
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		w.onSignal(sig)
	}()

	return &w
}
