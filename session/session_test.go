package session

import (
	"reflect"
	"testing"
)

// import "fmt"

func checkExported(i interface{}) (bool, []string) {
	exceptions := map[string]bool {"handleCommand" : true, "handleAction" : true, "quickRoom" : true }

	objType := reflect.TypeOf(i)

	pass := true
	var failedMethods []string

	for i := 0; i < objType.NumMethod(); i++ {
		methodType := objType.Method(i)

		_, found := exceptions[methodType.Name]

		if found {
			continue
		}

		if methodType.PkgPath != "" {
			pass = false
			failedMethods = append(failedMethods, methodType.Name)
		}
	}

	return pass, failedMethods
}

func Test_UserInputHandlers(t *testing.T) {
	args := make([]reflect.Value, 1)

	var strlist []string
	args[0] = reflect.ValueOf(strlist)

	var ah actionHandler
	var ch commandHandler

	passed, failedMethodNames := checkExported(&ah)
	if !passed {
		for _, failedMethodName := range failedMethodNames {
			t.Errorf("All methods should be exported: actionHandler." + failedMethodName)
		}
	}

	passed, failedMethodNames = checkExported(&ch)
	if !passed {
		for _, failedMethodName := range failedMethodNames {
			t.Errorf("All methods should be exported: commandHandler." + failedMethodName)
		}
	}
}

// vim:nocindent
