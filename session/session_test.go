package session

import (
	"reflect"
	"testing"
)

// Verify that all methods are exported (with some exceptions), and that they all
// take in a []string as their only argument
func checkMethods(i interface{}, t *testing.T) {
	objType := reflect.TypeOf(i)

	var stringArray []string
	stringArrayType := reflect.TypeOf(stringArray)

	for i := 0; i < objType.NumMethod(); i++ {
		methodValue := objType.Method(i)
		methodType := methodValue.Type

		structType := methodType.In(0)

		if methodValue.PkgPath != "" {
			continue
		}

		if methodType.NumIn() != 2 { // One for the receiver object, one for the actual argument
			t.Errorf("All methods must take exactly one parameter: %s.%s, %v", structType, methodValue.Name, methodType.NumIn())
			continue
		}

		paramType := methodType.In(1)

		if !stringArrayType.AssignableTo(paramType) {
			t.Errorf("All methods must take in a []string as their parameter: %s.%s", structType, methodValue.Name)
			continue
		}
	}
}

func Test_UserInputHandlers(t *testing.T) {
	args := make([]reflect.Value, 1)

	var strlist []string
	args[0] = reflect.ValueOf(strlist)

	var ah actionHandler

	checkMethods(&ah, t)
}

// vim:nocindent
