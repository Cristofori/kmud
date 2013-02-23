package telnet

import "testing"

func compareData(d1 []byte, d2 []byte) bool {
	if len(d1) != len(d2) {
		return false
	}

	for i, b := range d1 {
		if b != d2[i] {
			return false
		}
	}

	return true
}

func Test_Process(t *testing.T) {
	testStr := "test"

	data := []byte(testStr)

	result, subDataResult := Process(data)
	if result != "test" {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

    if subDataResult != nil {
        t.Errorf("Subdata should have been nil")
    }

	data = append(data, WillEcho()...)

	result, subDataResult = Process(data)
	if result != "test" {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

    if subDataResult != nil {
        t.Errorf("Subdata should have been nil")
    }

	data = append(data, []byte(" another test")...)

	result, subDataResult = Process(data)
	if result != "test another test" {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

    if subDataResult != nil {
        t.Errorf("Subdata should have been nil")
    }

	subData := []byte{'\x00', '\x12', '\x99'}

	data = append(data, buildCommand(SB, WS)...)
	data = append(data, subData...)
	data = append(data, buildCommand(SE)...)

	result, subDataResult = Process(data)
	if result != "test another test" {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if compareData(subDataResult[WS], subData) == false {
		t.Errorf("Process(%s), Subdata == %v, want %v", data, subDataResult[WS], subData)
	}

	data = append(data, []byte(" again")...)
	testStr = "test another test again"

	result, subDataResult = Process(data)
	if result != testStr {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if compareData(subDataResult[WS], subData) == false {
		t.Errorf("Process(%s), Subdata == %v, want %v", data, subDataResult[WS], subData)
	}

	subData = []byte{'\x00', '\x12', '\x99', '\xFF', '\xFF', '\x42'}
	wantedSubData := []byte{'\x00', '\x12', '\x99', '\xFF', '\x42'}

	data = append(data, buildCommand(SB, WS)...)
	data = append(data, subData...)
	data = append(data, buildCommand(SE)...)

	result, subDataResult = Process(data)
	if result != testStr {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if compareData(subDataResult[WS], wantedSubData) == false {
		t.Errorf("Process(%s), Subdata == %v, want %v", data, subDataResult[WS], wantedSubData)
	}

}

// vim: nocindent
