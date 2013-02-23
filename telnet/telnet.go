package telnet

import (
	"fmt"
)

// RFC 854: http://tools.ietf.org/html/rfc854, http://support.microsoft.com/kb/231866

var codeMap map[byte]TelnetCode
var commandMap map[TelnetCode]byte

type TelnetCode int

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
	if codeMap != nil {
		return
	}

	codeMap = map[byte]TelnetCode{}
	commandMap = map[TelnetCode]byte{}

	commandMap[NUL] = '\x00'
	commandMap[ECHO] = '\x01'
	commandMap[SGA] = '\x03'
	commandMap[ST] = '\x05'
	commandMap[TM] = '\x06'
	commandMap[BEL] = '\x07'
	commandMap[BS] = '\x08'
	commandMap[HT] = '\x09'
	commandMap[LF] = '\x0a'
	commandMap[FF] = '\x0c'
	commandMap[CR] = '\x0d'
	commandMap[TT] = '\x18'
	commandMap[WS] = '\x1F'
	commandMap[TS] = '\x20'
	commandMap[RFC] = '\x21'
	commandMap[LM] = '\x22'
	commandMap[EV] = '\x24'
	commandMap[SE] = '\xf0'
	commandMap[NOP] = '\xf1'
	commandMap[DM] = '\xf2'
	commandMap[BRK] = '\xf3'
	commandMap[IP] = '\xf4'
	commandMap[AO] = '\xf5'
	commandMap[AYT] = '\xf6'
	commandMap[EC] = '\xf7'
	commandMap[EL] = '\xf8'
	commandMap[GA] = '\xf9'
	commandMap[SB] = '\xfa'
	commandMap[WILL] = '\xfb'
	commandMap[WONT] = '\xfc'
	commandMap[DO] = '\xfd'
	commandMap[DONT] = '\xfe'
	commandMap[IAC] = '\xff'

	commandMap[CMP1] = '\x55'
	commandMap[CMP2] = '\x56'
	commandMap[AARD] = '\x66'
	commandMap[ATCP] = '\xc8'
	commandMap[GMCP] = '\xc9'

	for enum, code := range commandMap {
		codeMap[code] = enum
	}
}

// Process strips telnet control codes from the given input, returning the resulting input string
// and a map containing data of any subnegotiation parameters that were found
func Process(bytes []byte) (str string, subdata map[TelnetCode][]byte) {
	initLookups()

	subdata = map[TelnetCode][]byte{}

	type State int

	const (
		Base  State = iota
		InIAC State = iota
		InSB  State = iota
		CapSB State = iota
	)

	currentState := Base
	var currentSB TelnetCode

	capturedBytes := []byte{}

	capture := func(b byte) {
		capturedBytes = append(capturedBytes, b)
	}

	dontCapture := func(b byte) {
		str = str + string(b)
	}

	captureSubData := func(code TelnetCode, b byte) {
		subdata[code] = append(subdata[code], b)
	}

	skip := false

	for i, b := range bytes {
		if skip {
			skip = false
			continue
		}

		code := codeMap[b]

		switch currentState {
		case Base:
			if code == IAC {
				currentState = InIAC
				capture(b)
			} else {
				dontCapture(b)
			}

		case InIAC:
			if code == WILL || code == WONT || code == DO || code == DONT {
				// Stay in this state
			} else if code == SB {
				currentState = InSB
			} else {
				currentState = Base
			}
			capture(b)

		case InSB:
			capture(b)
			subdata[code] = []byte{}
			currentSB = code
			currentState = CapSB

		case CapSB:
			if code == IAC {
				// Check to see if it's an escaped IAC sequence (IAC IAC)
				if i+1 < len(bytes) {
					if codeMap[bytes[i+1]] == IAC {
						skip = true
						captureSubData(currentSB, b)
						continue
					}
				}

				currentState = InIAC
				capture(b)
			} else {
				captureSubData(currentSB, b)
			}
		}
	}

	if len(capturedBytes) > 0 {
		fmt.Println("Processed:", ToString(capturedBytes))
	}

	for key, value := range subdata {
		fmt.Printf("Subdata[%s] = %v\n", CodeToString(key), value)
	}

	return str, subdata
}

func ToString(bytes []byte) string {
	initLookups()

	str := ""
	for _, b := range bytes {

		if str != "" {
			str = str + " "
		}

		code, found := codeMap[b]

		if found {
			str = str + CodeToString(code)
		} else {
			str = str + "???"
		}
	}

	return str
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

	return "???"
}

func buildCommand(codes ...TelnetCode) []byte {
	command := make([]byte, len(codes)+1)
	command[0] = commandMap[IAC]

	for i, code := range codes {
		command[i+1] = commandMap[code]
	}

	return command
}

func WillEcho() []byte {
	return buildCommand(WILL, ECHO)
}

func WontEcho() []byte {
	return buildCommand(WONT, ECHO)
}

func DoWindowSize() []byte {
	return buildCommand(DO, WS)
}

// vim: nocindent
