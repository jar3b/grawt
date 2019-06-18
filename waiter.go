package grawt

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Waiter struct {
	blockingMode  bool
	waitGroup     sync.WaitGroup
	closeHandlers []*CloseHandler
}

func (w *Waiter) addHandler(f *func(), autoDone bool) *CloseHandler {
	ch := CloseHandler{
		w,
		make(chan struct{}, 1),
		true,
		autoDone,
		f,
	}
	w.waitGroup.Add(1)
	w.closeHandlers = append(w.closeHandlers, &ch)

	return &ch
}

func (w *Waiter) terminateHandler(h *CloseHandler, forceWaitGroupDone bool) {
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
	for _, h := range w.closeHandlers {
		if h.active {
			w.terminateHandler(h, false)
		}
	}
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
		false,
		sync.WaitGroup{},
		make([]*CloseHandler, 0),
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		w.onSignal(sig)
	}()

	return &w
}
