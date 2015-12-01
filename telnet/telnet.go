package telnet

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// RFC 854: http://tools.ietf.org/html/rfc854, http://support.microsoft.com/kb/231866

var byteToCode map[byte]TelnetCode
var codeToByte map[TelnetCode]byte

type TelnetCode int

// Telnet wraps the given connection object, processing telnet codes from its byte
// stream and interpreting them as necessary, making it possible to hand the connection
// object off to other code so that it doesn't have to worry about telnet escape sequences
// being found in its data.
type Telnet struct {
	conn net.Conn
	err  error

	processor *telnetProcessor
}

func NewTelnet(conn net.Conn) *Telnet {
	var t Telnet
	t.conn = conn
	t.processor = newTelnetProcessor()
	return &t
}

func (t *Telnet) Write(p []byte) (int, error) {
	return t.conn.Write(p)
}

func (t *Telnet) Read(p []byte) (int, error) {
	for {
		t.fill()
		if t.err != nil {
			return 0, t.err
		}

		n, err := t.processor.Read(p)
		if n > 0 {
			return n, err
		}
	}
}

func (t *Telnet) Data(code TelnetCode) []byte {
	return t.processor.subdata[code]
}

func (t *Telnet) Listen(listenFunc func(TelnetCode, []byte)) {
	t.processor.listenFunc = listenFunc
}

// Idea/name for this function shamelessly stolen from bufio
func (t *Telnet) fill() {
	buf := make([]byte, 1024)
	n, err := t.conn.Read(buf)
	t.err = err
	t.processor.addBytes(buf[:n])
}

func (t *Telnet) Close() error {
	return t.conn.Close()
}

func (t *Telnet) LocalAddr() net.Addr {
	return t.conn.LocalAddr()
}

func (t *Telnet) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}

func (t *Telnet) SetDeadline(dl time.Time) error {
	return t.conn.SetDeadline(dl)
}

func (t *Telnet) SetReadDeadline(dl time.Time) error {
	return t.conn.SetReadDeadline(dl)
}

func (t *Telnet) SetWriteDeadline(dl time.Time) error {
	return t.conn.SetWriteDeadline(dl)
}

func (t *Telnet) WillEcho() {
	t.SendCommand(WILL, ECHO)
}

func (t *Telnet) WontEcho() {
	t.SendCommand(WONT, ECHO)
}

func (t *Telnet) DoWindowSize() {
	t.SendCommand(DO, WS)
}

func (t *Telnet) DoTerminalType() {
	// This is really supposed to be two commands, one to ask if they'll send a
	// terminal type, and another to indicate that they should send it if
	// they've expressed a "willingness" to send it. For the time being this
	// works well enough.

	// See http://tools.ietf.org/html/rfc884

	t.SendCommand(DO, TT, IAC, SB, TT, 1, IAC, SE) // 1 = SEND
}

func (t *Telnet) SendCommand(codes ...TelnetCode) {
	t.conn.Write(BuildCommand(codes...))
}

func BuildCommand(codes ...TelnetCode) []byte {
	command := make([]byte, len(codes)+1)
	command[0] = codeToByte[IAC]

	for i, code := range codes {
		command[i+1] = codeToByte[code]
	}

	return command
}

const (
	NUL  TelnetCode = iota // NULL, no operation
	ECHO TelnetCode = iota // Echo
	SGA  TelnetCode = iota // Suppress go ahead
	ST   TelnetCode = iota // Status
	TM   TelnetCode = iota // Timing mark
	BEL  TelnetCode = iota // Bell
	BS   TelnetCode = iota // Backspace
	HT   TelnetCode = iota // Horizontal tab
	LF   TelnetCode = iota // Line feed
	FF   TelnetCode = iota // Form feed
	CR   TelnetCode = iota // Carriage return
	TT   TelnetCode = iota // Terminal type
	WS   TelnetCode = iota // Window size
	TS   TelnetCode = iota // Terminal speed
	RFC  TelnetCode = iota // Remote flow control
	LM   TelnetCode = iota // Line mode
	EV   TelnetCode = iota // Environment variables
	SE   TelnetCode = iota // End of subnegotiation parameters.
	NOP  TelnetCode = iota // No operation.
	DM   TelnetCode = iota // Data Mark. The data stream portion of a Synch. This should always be accompanied by a TCP Urgent notification.
	BRK  TelnetCode = iota // Break. NVT character BRK.
	IP   TelnetCode = iota // Interrupt Process
	AO   TelnetCode = iota // Abort output
	AYT  TelnetCode = iota // Are you there
	EC   TelnetCode = iota // Erase character
	EL   TelnetCode = iota // Erase line
	GA   TelnetCode = iota // Go ahead signal
	SB   TelnetCode = iota // Indicates that what follows is subnegotiation of the indicated option.
	WILL TelnetCode = iota // Indicates the desire to begin performing, or confirmation that you are now performing, the indicated option.
	WONT TelnetCode = iota // Indicates the refusal to perform, or continue performing, the indicated option.
	DO   TelnetCode = iota // Indicates the request that the other party perform, or confirmation that you are expecting the other party to perform, the indicated option.
	DONT TelnetCode = iota // Indicates the demand that the other party stop performing, or confirmation that you are no longer expecting the other party to perform, the indicated option.
	IAC  TelnetCode = iota // Interpret as command

	// Non-standard codes:
	CMP1 TelnetCode = iota // MCCP Compress
	CMP2 TelnetCode = iota // MCCP Compress2
	AARD TelnetCode = iota // Aardwolf MUD out of band communication, http://www.aardwolf.com/blog/2008/07/10/telnet-negotiation-control-mud-client-interaction/
	ATCP TelnetCode = iota // Achaea Telnet Client Protocol, http://www.ironrealms.com/rapture/manual/files/FeatATCP-txt.html
	GMCP TelnetCode = iota // Generic Mud Communication Protocol
)

func initLookups() {
	if byteToCode != nil {
		return
	}

	byteToCode = map[byte]TelnetCode{}
	codeToByte = map[TelnetCode]byte{}

	codeToByte[NUL] = '\x00'
	codeToByte[ECHO] = '\x01'
	codeToByte[SGA] = '\x03'
	codeToByte[ST] = '\x05'
	codeToByte[TM] = '\x06'
	codeToByte[BEL] = '\x07'
	codeToByte[BS] = '\x08'
	codeToByte[HT] = '\x09'
	codeToByte[LF] = '\x0a'
	codeToByte[FF] = '\x0c'
	codeToByte[CR] = '\x0d'
	codeToByte[TT] = '\x18'
	codeToByte[WS] = '\x1F'
	codeToByte[TS] = '\x20'
	codeToByte[RFC] = '\x21'
	codeToByte[LM] = '\x22'
	codeToByte[EV] = '\x24'
	codeToByte[SE] = '\xf0'
	codeToByte[NOP] = '\xf1'
	codeToByte[DM] = '\xf2'
	codeToByte[BRK] = '\xf3'
	codeToByte[IP] = '\xf4'
	codeToByte[AO] = '\xf5'
	codeToByte[AYT] = '\xf6'
	codeToByte[EC] = '\xf7'
	codeToByte[EL] = '\xf8'
	codeToByte[GA] = '\xf9'
	codeToByte[SB] = '\xfa'
	codeToByte[WILL] = '\xfb'
	codeToByte[WONT] = '\xfc'
	codeToByte[DO] = '\xfd'
	codeToByte[DONT] = '\xfe'
	codeToByte[IAC] = '\xff'

	codeToByte[CMP1] = '\x55'
	codeToByte[CMP2] = '\x56'
	codeToByte[AARD] = '\x66'
	codeToByte[ATCP] = '\xc8'
	codeToByte[GMCP] = '\xc9'

	for enum, code := range codeToByte {
		byteToCode[code] = enum
	}
}

type processorState int

const (
	stateBase   processorState = iota
	stateInIAC  processorState = iota
	stateInSB   processorState = iota
	stateCapSB  processorState = iota
	stateEscIAC processorState = iota
)

// telnetProcessor implements a state machine that reads input one byte at a time
// and processes it according to the telnet spec. It is designed to read a raw telnet
// stream, from which it will extract telnet escape codes and subnegotiation data.
// The processor can then be read from with all of the telnet codes removed, leaving
// the pure user input stream.
type telnetProcessor struct {
	state     processorState
	currentSB TelnetCode

	capturedBytes []byte
	subdata       map[TelnetCode][]byte
	cleanData     string
	listenFunc    func(TelnetCode, []byte)

	debug bool
}

func newTelnetProcessor() *telnetProcessor {
	initLookups()

	var tp telnetProcessor
	tp.state = stateBase
	tp.debug = false
	tp.currentSB = NUL

	return &tp
}

func (self *telnetProcessor) Read(p []byte) (int, error) {
	maxLen := len(p)

	n := 0

	if maxLen >= len(self.cleanData) {
		n = len(self.cleanData)
	} else {
		n = maxLen
	}

	for i := 0; i < n; i++ {
		p[i] = self.cleanData[i]
	}

	self.cleanData = self.cleanData[n:] // TODO: Memory leak?

	return n, nil
}

func (self *telnetProcessor) capture(b byte) {
	if self.debug {
		fmt.Println("Captured:", ByteToCodeString(b))
	}

	self.capturedBytes = append(self.capturedBytes, b)
}

func (self *telnetProcessor) dontCapture(b byte) {
	self.cleanData = self.cleanData + string(b)
}

func (self *telnetProcessor) resetSubDataField(code TelnetCode) {
	if self.subdata == nil {
		self.subdata = map[TelnetCode][]byte{}
	}

	self.subdata[code] = []byte{}
}

func (self *telnetProcessor) captureSubData(code TelnetCode, b byte) {
	if self.debug {
		fmt.Println("Captured subdata:", CodeToString(code), b)
	}

	if self.subdata == nil {
		self.subdata = map[TelnetCode][]byte{}
	}

	self.subdata[code] = append(self.subdata[code], b)
}

func (self *telnetProcessor) addBytes(bytes []byte) {
	for _, b := range bytes {
		self.addByte(b)
	}
}

func (self *telnetProcessor) addByte(b byte) {
	code := byteToCode[b]

	switch self.state {
	case stateBase:
		if code == IAC {
			self.state = stateInIAC
			self.capture(b)
		} else {
			self.dontCapture(b)
		}

	case stateInIAC:
		if code == WILL || code == WONT || code == DO || code == DONT {
			// Stay in this state
		} else if code == SB {
			self.state = stateInSB
		} else {
			self.state = stateBase
		}
		self.capture(b)

	case stateInSB:
		self.capture(b)
		self.currentSB = code
		self.state = stateCapSB
		self.resetSubDataField(code)

	case stateCapSB:
		if code == IAC {
			self.state = stateEscIAC
		} else {
			self.captureSubData(self.currentSB, b)
		}

	case stateEscIAC:
		if code == IAC {
			self.state = stateCapSB
			self.captureSubData(self.currentSB, b)
		} else {
			self.subDataFinished(self.currentSB)
			self.currentSB = NUL
			self.state = stateBase
			self.addByte(codeToByte[IAC])
			self.addByte(b)
		}
	}
}

func (self *telnetProcessor) subDataFinished(code TelnetCode) {
	if self.listenFunc != nil {
		self.listenFunc(code, self.subdata[code])
	}
}

func ToString(bytes []byte) string {
	initLookups()

	str := ""
	for _, b := range bytes {

		if str != "" {
			str = str + " "
		}

		str = str + ByteToCodeString(b)
	}

	return str
}

func ByteToCodeString(b byte) string {
	code, found := byteToCode[b]

	if !found {
		return "??(" + strconv.Itoa(int(b)) + ")"
	}

	return CodeToString(code)
}

func CodeToString(code TelnetCode) string {
	switch code {
	case NUL:
		return "NUL"
	case ECHO:
		return "ECHO"
	case SGA:
		return "SGA"
	case ST:
		return "ST"
	case TM:
		return "TM"
	case BEL:
		return "BEL"
	case BS:
		return "BS"
	case HT:
		return "HT"
	case LF:
		return "LF"
	case FF:
		return "FF"
	case CR:
		return "CR"
	case TT:
		return "TT"
	case WS:
		return "WS"
	case TS:
		return "TS"
	case RFC:
		return "RFC"
	case LM:
		return "LM"
	case EV:
		return "EV"
	case SE:
		return "SE"
	case NOP:
		return "NOP"
	case DM:
		return "DM"
	case BRK:
		return "BRK"
	case IP:
		return "IP"
	case AO:
		return "AO"
	case AYT:
		return "AYT"
	case EC:
		return "EC"
	case EL:
		return "EL"
	case GA:
		return "GA"
	case SB:
		return "SB"
	case WILL:
		return "WILL"
	case WONT:
		return "WONT"
	case DO:
		return "DO"
	case DONT:
		return "DONT"
	case IAC:
		return "IAC"
	case CMP1:
		return "CMP1"
	case CMP2:
		return "CMP2"
	case AARD:
		return "AARD"
	case ATCP:
		return "ATCP"
	case GMCP:
		return "GMCP"
	}

	return ""
}
