package telnet

import (
	"fmt"
)

// RFC 854: http://tools.ietf.org/html/rfc854, http://support.microsoft.com/kb/231866

var codeMap map[byte]int
var commandMap map[int]byte

const (
	NUL  = iota // NULL, no operation
	ECHO = iota // Echo
	SGA  = iota // Suppress go ahead
	ST   = iota // Status
	TM   = iota // Timing mark
	BEL  = iota // Bell
	BS   = iota // Backspace
	HT   = iota // Horizontal tab
	LF   = iota // Line feed
	FF   = iota // Form feed
	CR   = iota // Carriage return
	TT   = iota // Terminal type
	WS   = iota // Window size
	TS   = iota // Terminal speed
	RFC  = iota // Remote flow control
	LM   = iota // Line mode
	EV   = iota // Environment variables
	SE   = iota // End of subnegotiation parameters.
	NOP  = iota // No operation.
	DM   = iota // Data Mark. The data stream portion of a Synch. This should always be accompanied by a TCP Urgent notification.
	BRK  = iota // Break. NVT character BRK.
	IP   = iota // Interrupt Process
	AO   = iota // Abort output
	AYT  = iota // Are you there
	EC   = iota // Erase character
	EL   = iota // Erase line
	GA   = iota // Go ahead signal
	SB   = iota // Indicates that what follows is subnegotiation of the indicated option.
	WILL = iota // Indicates the desire to begin performing, or confirmation that you are now performing, the indicated option.
	WONT = iota // Indicates the refusal to perform, or continue performing, the indicated option.
	DO   = iota // Indicates the request that the other party perform, or confirmation that you are expecting the other party to perform, the indicated option.
	DONT = iota // Indicates the demand that the other party stop performing, or confirmation that you are no longer expecting the other party to perform, the indicated option.
	IAC  = iota // Interpret as command

	// Non-standard codes:
	CMP1 = iota // MCCP Compress
	CMP2 = iota // MCCP Compress2
	AARD = iota // Aardwolf MUD out of band communication, http://www.aardwolf.com/blog/2008/07/10/telnet-negotiation-control-mud-client-interaction/
	ATCP = iota // Achaea Telnet Client Protocol, http://www.ironrealms.com/rapture/manual/files/FeatATCP-txt.html
	GMCP = iota // Generic Mud Communication Protocol
)

func initLookups() {
	if codeMap != nil {
		return
	}

	codeMap = map[byte]int{}
	commandMap = map[int]byte{}

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
func Process(bytes []byte) string {
	initLookups()

	str := ""
	var bytesProcessed []byte

	inIAC := false

	processByte := func(b byte) {
		bytesProcessed = append(bytesProcessed, b)
	}

	for _, b := range bytes {
		if b == commandMap[IAC] {
			inIAC = true
			processByte(b)
			continue
		}

		if inIAC {
			if b != commandMap[WILL] && b != commandMap[WONT] && b != commandMap[DO] && b != commandMap[DONT] {
				inIAC = false
			}
			processByte(b)
			continue
		}

		str = str + string(b)
	}

	if len(bytesProcessed) > 0 {
		fmt.Printf("Processed: %s\n", ToString(bytesProcessed))
	}

	return str
}

func Code(enum int) byte {
	initLookups()
	return commandMap[enum]
}

func ToString(bytes []byte) string {
	initLookups()

	str := ""
	for _, b := range bytes {
		enum, found := codeMap[b]
		result := ""

		if found {
			switch enum {
			case NUL:
				result = "NUL"
			case ECHO:
				result = "ECHO"
			case SGA:
				result = "SGA"
			case ST:
				result = "ST"
			case TM:
				result = "TM"
			case BEL:
				result = "BEL"
			case BS:
				result = "BS"
			case HT:
				result = "HT"
			case LF:
				result = "LF"
			case FF:
				result = "FF"
			case CR:
				result = "CR"
			case TT:
				result = "TT"
			case WS:
				result = "WS"
			case TS:
				result = "TS"
			case RFC:
				result = "RFC"
			case LM:
				result = "LM"
			case EV:
				result = "EV"
			case SE:
				result = "SE"
			case NOP:
				result = "NOP"
			case DM:
				result = "DM"
			case BRK:
				result = "BRK"
			case IP:
				result = "IP"
			case AO:
				result = "AO"
			case AYT:
				result = "AYT"
			case EC:
				result = "EC"
			case EL:
				result = "EL"
			case GA:
				result = "GA"
			case SB:
				result = "SB"
			case WILL:
				result = "WILL"
			case WONT:
				result = "WONT"
			case DO:
				result = "DO"
			case DONT:
				result = "DONT"
			case IAC:
				result = "IAC"
			case CMP1:
				result = "CMP1"
			case CMP2:
				result = "CMP2"
			case AARD:
				result = "AARD"
			case ATCP:
				result = "ATCP"
			case GMCP:
				result = "GMCP"
			}
		} else {
			result = "???"
		}

		if str != "" {
			str = str + " "
		}
		str = str + result
	}

	return str
}

func buildCommand(length int) []byte {
	command := make([]byte, length)
	command[0] = commandMap[IAC]
	return command
}

func WillEcho() []byte {
	command := buildCommand(3)
	command[1] = commandMap[WILL]
	command[2] = commandMap[ECHO]
	return command
}

func WontEcho() []byte {
	command := buildCommand(3)
	command[1] = commandMap[WONT]
	command[2] = commandMap[ECHO]
	return command
}

// vim: nocindent
