# grawt
graceful terminations for go applications

## example usage

```
var waiter = grawt.NewWaiter()

waiter.AddCloseHandler(func() {
		nacl.FinalizeStan()
	}, false)

// blocking wait, if no need to block (with http server, for example), you can omit .Wait() call
waiter.Wait()
```