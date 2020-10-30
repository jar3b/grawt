package grawt

type CloseHandler struct {
	waiter      *Waiter
	Quit        chan struct{}
	active      bool
	autoDone    bool
	handlerFunc *func()
}

func (ch *CloseHandler) Halt(err error) {
	ch.waiter.Halt(err)
}
