package engine

var _listeners map[*chan interface{}]bool

func Register() *chan interface{} {
	if _listeners == nil {
		_listeners = map[*chan interface{}]bool{}
	}

	listener := make(chan interface{}, 100)
	_listeners[&listener] = true
	return &listener
}

func Unregister(listener *chan interface{}) {
	delete(_listeners, listener)
}

func broadcast(event interface{}) {
	for listener, _ := range _listeners {
		*listener <- event
	}
}
