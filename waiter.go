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
	isHalting     bool
}

func (w *Waiter) addHandler(f *func(), autoDone bool) *CloseHandler {
	ch := CloseHandler{
		waiter:      w,
		Quit:        make(chan struct{}, 1),
		active:      true,
		autoDone:    autoDone,
		handlerFunc: f,
	}

	w.Lock()
	w.waitGroup.Add(1)
	w.closeHandlers = append(w.closeHandlers, &ch)
	w.Unlock()

	return &ch
}

func (w *Waiter) terminateHandler(h *CloseHandler, forceWaitGroupDone bool) {
	if !h.active {
		return
	}
	if h.handlerFunc != nil && *h.handlerFunc != nil {
		(*h.handlerFunc)()
	}

	if h.active {
		close(h.Quit)
	}
	if h.autoDone || forceWaitGroupDone {
		w.waitGroup.Done()
	}

	h.active = false
}

func (w *Waiter) AddCloseHandler(f func(), waitForChannel bool) *CloseHandler {
	return w.addHandler(&f, !waitForChannel)
}

func (w *Waiter) Halt(err error) {
	if w.isHalting {
		return
	}
	w.isHalting = true

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

	w.isHalting = false
}

func (w *Waiter) Wait() {
	w.blockingMode = true
	log.Info("Waiting...")
	w.waitGroup.Wait()
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
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		w.onSignal(sig)
	}()

	return &w
}
