package testutils

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

type TestWriter struct {
	Wrote string
}

func (self *TestWriter) Write(p []byte) (n int, err error) {
	self.Wrote += string(p)
	return len(p), nil
}

type TestReader struct {
	ToRead string
	err    error
}

func (self *TestReader) Read(p []byte) (n int, err error) {
	if self.err != nil {
		return 0, self.err
	}

	for i := 0; i < len(self.ToRead); i++ {
		p[i] = self.ToRead[i]
	}

	p[len(self.ToRead)] = '\n'

	return len(self.ToRead) + 1, nil
}

func (self *TestReader) SetError(err error) {
	self.err = err
}

type TestReadWriter struct {
	TestReader
	TestWriter
}

type TestCommunicable struct {
	TestReadWriter
}

func (self *TestCommunicable) Write(text string) {
	self.TestReadWriter.Write([]byte(text))
}

func (self *TestCommunicable) WriteLine(line string) {
	self.Write(line + "\n")
}

func (self *TestCommunicable) GetInput(prompt string) string {
	self.Write(prompt)
	buf := make([]byte, 1024)
	n, _ := self.Read(buf)
	return strings.ToLower(strings.TrimSpace(string(buf[:n])))
}

func (self *TestCommunicable) GetWindowSize() (int, int) {
	return 80, 40
}

func TestSettersAndGetters(object interface{}, t *testing.T) bool {
	objType := reflect.TypeOf(object)

	regex, _ := regexp.Compile("^Get(.+)")

	getterToSetter := make(map[string]string)

	for i := 0; i < objType.NumMethod(); i++ {
		method := objType.Method(i)

		findMatchingFunctions := func(prefix1, prefix2 string) string {
			if strings.HasPrefix(method.Name, prefix1) {
				result := regex.FindStringSubmatch(method.Name)

				if result != nil {
					pairName := "Set" + result[1]
					_, found := objType.MethodByName(pairName)

					if !found {
						t.Logf("Unable to find matching setter/getter for %s.%s", objType.String(), method.Name)
						return ""
					}

					return pairName
				}
			}

			return ""
		}

		pairedMethodName := findMatchingFunctions("Get", "Set")
		if pairedMethodName != "" {
			getterToSetter[method.Name] = pairedMethodName
		}

		findMatchingFunctions("Set", "Get")
	}

	v := reflect.ValueOf(object)

	for g, s := range getterToSetter {
		getterValue := v.MethodByName(g)
		setterValue := v.MethodByName(s)

		getterType := getterValue.Type()
		setterType := setterValue.Type()

		if getterType.NumOut() != setterType.NumIn() {
			t.Errorf("In/out mismatch: %s:%v, %s:%v", g, getterType.NumOut(), s, setterType.NumIn())
		} else {
			vals := make([]reflect.Value, setterType.NumIn())

			for i := 0; i < len(vals); i++ {
				inType := setterType.In(i)
				t.Log("inType:", inType)
				vals[i] = reflect.New(inType)
			}

			setterValue.Call(vals)
		}
	}

	return true
}

func Assert(condition bool, t *testing.T, failMessage ...interface{}) {
	if !condition {
		t.Error(failMessage...)
	}
}
