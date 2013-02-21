package session

var actionMap map[string]func([]string)

func registerAction(action string, actionFunc func([]string)) {
	if actionMap == nil {
		actionMap = map[string]func([]string){}
	}

	actionMap[action] = actionFunc
}

func registerActions(actions []string, actionFunc func([]string)) {
	for _, action := range actions {
		registerAction(action, actionFunc)
	}
}

func callAction(action string, args []string) bool {
	actionFunc, found := actionMap[action]

	if !found {
		return false
	}

	actionFunc(args)
	return true
}

func makeList(argList ...string) []string {
	list := make([]string, len(argList))
	for i, arg := range argList {
		list[i] = arg
	}
	return list
}

// vim: nocindent
