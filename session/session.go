package session

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/Cristofori/kmud/combat"
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	// "log"
	// "os"
	"strings"
	"time"
)

type Session struct {
	conn io.ReadWriter
	user types.User
	pc   types.PC

	prompt string
	states map[string]string

	userInputChannel chan string
	inputModeChannel chan userInputMode
	prompterChannel  chan utils.Prompter
	panicChannel     chan interface{}
	eventChannel     chan events.Event

	silentMode bool
	replyId    types.Id
	lastInput  string

	// logger *log.Logger
}

func NewSession(conn io.ReadWriter, user types.User, pc types.PC) *Session {
	var session Session
	session.conn = conn
	session.user = user
	session.pc = pc

	session.prompt = "%h/%H> "
	session.states = map[string]string{}

	session.userInputChannel = make(chan string)
	session.inputModeChannel = make(chan userInputMode)
	session.prompterChannel = make(chan utils.Prompter)
	session.panicChannel = make(chan interface{})
	session.eventChannel = events.Register(pc)

	session.silentMode = false

	// file, err := os.OpenFile(pc.GetName()+".log", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	// utils.PanicIfError(err)

	// session.logger = log.New(file, pc.GetName()+" ", log.LstdFlags)

	model.Login(pc)

	return &session
}

type userInputMode int

const (
	CleanUserInput userInputMode = iota
	RawUserInput   userInputMode = iota
)

func (self *Session) Exec() {
	defer events.Unregister(self.pc)
	defer model.Logout(self.pc)

	self.WriteLine("Welcome, " + self.pc.GetName())
	self.PrintRoom()

	// Main routine in charge of actually reading input from the connection object,
	// also has built in throttling to limit how fast we are allowed to process
	// commands from the user.
	go func() {
		defer func() {
			self.panicChannel <- recover()
		}()

		throttler := utils.NewThrottler(200 * time.Millisecond)

		for {
			mode := <-self.inputModeChannel
			prompter := <-self.prompterChannel
			input := ""

			switch mode {
			case CleanUserInput:
				input = utils.GetUserInputP(self.conn, prompter, self.user.GetColorMode())
			case RawUserInput:
				input = utils.GetRawUserInputP(self.conn, prompter, self.user.GetColorMode())
			default:
				panic("Unhandled case in switch statement (userInputMode)")
			}

			throttler.Sync()
			self.userInputChannel <- input
		}
	}()

	// Main loop
	for {
		input := self.getUserInputP(RawUserInput, self)
		if input == "" || input == "logout" || input == "quit" {
			return
		}

		if input == "." {
			input = self.lastInput
		}

		self.lastInput = input

		if strings.HasPrefix(input, "/") {
			self.handleCommand(utils.Argify(input[1:]))
		} else {
			self.handleAction(utils.Argify(input))
		}
	}
}

func (self *Session) WriteLinef(line string, a ...interface{}) {
	self.WriteLineColor(types.ColorWhite, line, a...)
}

func (self *Session) WriteLine(line string, a ...interface{}) {
	self.WriteLineColor(types.ColorWhite, line, a...)
}

func (self *Session) WriteLineColor(color types.Color, line string, a ...interface{}) {
	self.printLine(types.Colorize(color, fmt.Sprintf(line, a...)))
}

func (self *Session) printLine(line string, a ...interface{}) {
	self.Write(fmt.Sprintf(line+"\r\n", a...))
}

func (self *Session) Write(text string) {
	self.user.Write(text)
}

func (self *Session) printError(err string, a ...interface{}) {
	self.WriteLineColor(types.ColorRed, err, a...)
}

func (self *Session) clearLine() {
	utils.ClearLine(self.conn)
}

func (self *Session) asyncMessage(message string) {
	self.clearLine()
	self.WriteLine(message)
}

func (self *Session) GetInput(prompt string) string {
	return self.getUserInput(CleanUserInput, prompt)
}

func (self *Session) GetWindowSize() (int, int) {
	return self.user.GetWindowSize()
}

// getUserInput allows us to retrieve user input in a way that doesn't block the
// event loop by using channels and a separate Go routine to grab
// either the next user input or the next event.
func (self *Session) getUserInputP(inputMode userInputMode, prompter utils.Prompter) string {
	self.inputModeChannel <- inputMode
	self.prompterChannel <- prompter

	for {
		select {
		case input := <-self.userInputChannel:
			return input
		case event := <-self.eventChannel:
			if self.silentMode {
				continue
			}

			switch e := event.(type) {
			case events.TellEvent:
				self.replyId = e.From.GetId()
			case events.TickEvent:
				if !combat.InCombat(self.pc) {
					oldHps := self.pc.GetHitPoints()
					self.pc.Heal(5)
					newHps := self.pc.GetHitPoints()

					if oldHps != newHps {
						self.clearLine()
						self.Write(prompter.GetPrompt())
					}
				}
			}

			message := event.ToString(self.pc)
			if message != "" {
				self.asyncMessage(message)
				self.Write(prompter.GetPrompt())
			}

		case quitMessage := <-self.panicChannel:
			panic(quitMessage)
		}
	}
}

func (self *Session) getUserInput(inputMode userInputMode, prompt string) string {
	return self.getUserInputP(inputMode, utils.SimplePrompter(prompt))
}

func (self *Session) getCleanUserInput(prompt string) string {
	return self.getUserInput(CleanUserInput, prompt)
}

func (self *Session) getRawUserInput(prompt string) string {
	return self.getUserInput(RawUserInput, prompt)
}

func (self *Session) getConfirmation(prompt string) bool {
	answer := self.getCleanUserInput(types.Colorize(types.ColorWhite, prompt))
	return answer == "y" || answer == "yes"
}

func (self *Session) getInt(prompt string, min, max int) (int, bool) {
	for {
		input := self.getRawUserInput(prompt)
		if input == "" {
			return 0, false
		}

		val, err := utils.Atoir(input, min, max)

		if err != nil {
			self.printError(err.Error())
		} else {
			return val, true
		}
	}
}

func (self *Session) GetPrompt() string {
	prompt := self.prompt
	prompt = strings.Replace(prompt, "%h", strconv.Itoa(self.pc.GetHitPoints()), -1)
	prompt = strings.Replace(prompt, "%H", strconv.Itoa(self.pc.GetHealth()), -1)

	if len(self.states) > 0 {
		states := make([]string, len(self.states))

		i := 0
		for key, value := range self.states {
			states[i] = fmt.Sprintf("%s:%s", key, value)
			i++
		}

		prompt = fmt.Sprintf("%s %s", states, prompt)
	}

	return types.Colorize(types.ColorWhite, prompt)
}

func (self *Session) currentZone() types.Zone {
	return model.GetZone(self.GetRoom().GetZoneId())
}

func (self *Session) handleAction(action string, arg string) {
	if arg == "" {
		direction := types.StringToDirection(action)

		if direction != types.DirectionNone {
			if self.GetRoom().HasExit(direction) {
				err := model.MoveCharacter(self.pc, direction)
				if err == nil {
					self.PrintRoom()
				} else {
					self.printError(err.Error())
				}

			} else {
				self.printError("You can't go that way")
			}

			return
		}
	}

	handler, found := actions[action]

	if found {
		if handler.alias != "" {
			handler = actions[handler.alias]
		}
		handler.exec(self, arg)
	} else {
		self.printError("You can't do that")
	}
}

func (self *Session) handleCommand(name string, arg string) {
	if len(name) == 0 {
		return
	}

	if name[0] == '/' && self.user.IsAdmin() {
		quickRoom(self, name[1:])
		return
	}

	command, found := commands[name]

	if found {
		if command.alias != "" {
			command = commands[command.alias]
		}

		if command.admin && !self.user.IsAdmin() {
			self.printError("You don't have permission to do that")
		} else {
			command.exec(command, self, arg)
		}
	} else {
		self.printError("Unrecognized command: %s", name)
	}
}

func (self *Session) GetRoom() types.Room {
	return model.GetRoom(self.pc.GetRoomId())
}

func (self *Session) PrintRoom() {
	self.printRoom(self.GetRoom())
}

func (self *Session) printRoom(room types.Room) {
	pcs := model.PlayerCharactersIn(self.pc.GetRoomId(), self.pc.GetId())
	npcs := model.NpcsIn(room.GetId())
	items := model.ItemsIn(room.GetId())
	store := model.StoreIn(room.GetId())

	var area types.Area
	if room.GetAreaId() != nil {
		area = model.GetArea(room.GetAreaId())
	}

	var str string

	areaStr := ""
	if area != nil {
		areaStr = fmt.Sprintf("%s - ", area.GetName())
	}

	str = fmt.Sprintf("\r\n %v>>> %v%s%s %v<<< %v(%v %v %v)\r\n\r\n %v%s\r\n\r\n",
		types.ColorWhite, types.ColorBlue,
		areaStr, room.GetTitle(),
		types.ColorWhite, types.ColorBlue,
		room.GetLocation().X, room.GetLocation().Y, room.GetLocation().Z,
		types.ColorWhite,
		room.GetDescription())

	if store != nil {
		str = fmt.Sprintf("%s Store: %s\r\n\r\n", str, types.Colorize(types.ColorBlue, store.GetName()))
	}

	extraNewLine := ""

	if len(pcs) > 0 {
		str = fmt.Sprintf("%s %sAlso here:", str, types.ColorBlue)

		names := make([]string, len(pcs))
		for i, char := range pcs {
			names[i] = types.Colorize(types.ColorWhite, char.GetName())
		}
		str = fmt.Sprintf("%s %s \r\n", str, strings.Join(names, types.Colorize(types.ColorBlue, ", ")))

		extraNewLine = "\r\n"
	}

	if len(npcs) > 0 {
		str = fmt.Sprintf("%s %s", str, types.Colorize(types.ColorBlue, "NPCs: "))

		names := make([]string, len(npcs))
		for i, npc := range npcs {
			names[i] = types.Colorize(types.ColorWhite, npc.GetName())
		}
		str = fmt.Sprintf("%s %s \r\n", str, strings.Join(names, types.Colorize(types.ColorBlue, ", ")))

		extraNewLine = "\r\n"
	}

	if len(items) > 0 {
		itemMap := make(map[string]int)
		var nameList []string

		for _, item := range items {
			if item == nil {
				continue
			}

			_, found := itemMap[item.GetName()]
			if !found {
				nameList = append(nameList, item.GetName())
			}
			itemMap[item.GetName()]++
		}

		sort.Strings(nameList)

		str = str + " " + types.Colorize(types.ColorBlue, "Items: ")

		var names []string
		for _, name := range nameList {
			if itemMap[name] > 1 {
				name = fmt.Sprintf("%s x%v", name, itemMap[name])
			}
			names = append(names, types.Colorize(types.ColorWhite, name))
		}
		str = str + strings.Join(names, types.Colorize(types.ColorBlue, ", ")) + "\r\n"

		extraNewLine = "\r\n"
	}

	str = str + extraNewLine + " " + types.Colorize(types.ColorBlue, "Exits: ")

	var exitList []string
	for _, direction := range room.GetExits() {
		exitList = append(exitList, utils.DirectionToExitString(direction))
	}

	if len(exitList) == 0 {
		str = str + types.Colorize(types.ColorWhite, "None")
	} else {
		str = str + strings.Join(exitList, " ")
	}

	if len(room.GetLinks()) > 0 {
		str = fmt.Sprintf("%s\r\n\r\n %s %s",
			str,
			types.Colorize(types.ColorBlue, "Other exits:"),
			types.Colorize(types.ColorWhite, strings.Join(room.LinkNames(), ", ")),
		)
	}

	str = str + "\r\n"

	self.WriteLine(str)
}
