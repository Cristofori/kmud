package telnet

import (
	"bufio"
	"net"
	"testing"
	"time"
)

type fakeConn struct {
	data []byte
}

func (self *fakeConn) Write(p []byte) (int, error) {
	self.data = append(self.data, p...)
	return len(p), nil
}

func (self *fakeConn) Read(p []byte) (int, error) {
	n := 0

	for i := 0; i < len(p) && i < len(self.data); i++ {
		p[i] = self.data[i]
		n++
	}

	self.data = self.data[n:]

	return n, nil
}

func (self *fakeConn) Close() error {
	return nil
}

func (self *fakeConn) LocalAddr() net.Addr {
	return nil
}

func (self *fakeConn) RemoteAddr() net.Addr {
	return nil
}

func (self *fakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (self *fakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (self *fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}

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

func Test_Processor(t *testing.T) {
	var fc fakeConn
	fc.data = []byte{}

	telnet := NewTelnet(&fc)
	testStr := "test"
	readBuffer := make([]byte, 1024)

	data := []byte(testStr)
	telnet.Write(data)
	n, err := telnet.Read(readBuffer)
	result := readBuffer[:n]

	if compareData(result, []byte(testStr)) == false || err != nil {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	subdataResult := telnet.Data(WS)
	if subdataResult != nil {
		t.Errorf("Subdata should have been nil")
	}

	data = append(data, BuildCommand(WILL, ECHO)...)
	telnet.Write(data)
	n, err = telnet.Read(readBuffer)
	result = readBuffer[:n]

	if compareData(result, []byte(testStr)) == false || err != nil {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if telnet.processor.subdata != nil {
		t.Errorf("Subdata should have been nil")
	}

	data = append(data, []byte(" another test")...)
	testStr = testStr + " another test"
	telnet.Write(data)
	n, err = telnet.Read(readBuffer)
	result = readBuffer[:n]

	if compareData(result, []byte(testStr)) == false {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if telnet.processor.subdata != nil {
		t.Errorf("Subdata should have been nil")
	}

	subData := []byte{'\x00', '\x12', '\x99'}

	data = append(data, BuildCommand(SB, WS)...)
	data = append(data, subData...)
	data = append(data, BuildCommand(SE)...)

	telnet.Write(data)
	n, err = telnet.Read(readBuffer)
	result = readBuffer[:n]

	if compareData(result, []byte(testStr)) == false {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if compareData(telnet.Data(WS), subData) == false {
		t.Errorf("Process(%s), Subdata == %v, want %v", data, telnet.Data(WS), subData)
	}

	data = append(data, []byte(" again")...)
	testStr = testStr + " again"

	telnet.Write(data)
	n, err = telnet.Read(readBuffer)
	result = readBuffer[:n]

	if compareData(result, []byte(testStr)) == false {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if compareData(telnet.Data(WS), subData) == false {
		t.Errorf("Process(%s), Subdata == %v, want %v", data, telnet.Data(WS), subData)
	}

	// Interpret escaped FF bytes properly
	subData = []byte{'\x00', '\x12', '\x99', '\xFF', '\xFF', '\x42'}
	wantedSubData := []byte{'\x00', '\x12', '\x99', '\xFF', '\x42'}

	data = append(data, BuildCommand(SB, WS)...)
	data = append(data, subData...)
	data = append(data, BuildCommand(SE)...)

	telnet.Write(data)
	n, err = telnet.Read(readBuffer)
	result = readBuffer[:n]

	if compareData(result, []byte(testStr)) == false {
		t.Errorf("Process(%s) == '%s', want '%s'", data, result, testStr)
	}

	if compareData(telnet.Data(WS), wantedSubData) == false {
		t.Errorf("Process(%s), Subdata == %v, want %v", data, telnet.Data(WS), wantedSubData)
	}

	// Test with bufio
	testStr = "bufio test\n"
	data = []byte(testStr)

	reader := bufio.NewReader(telnet)
	telnet.Write(data)

	bytes, err := reader.ReadBytes('\n')

	if compareData(bytes, data) == false {
		t.Errorf("Bufio failure %v != %v", bytes, data)
	}
}
