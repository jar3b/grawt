# grawt
graceful terminations for go applications

## example usage

```
var waiter = grawt.NewWaiter()

waiter.AddCloseHandler(func() {
		nacl.FinalizeStan()
	}, false)

waiter.Wait(true)
```