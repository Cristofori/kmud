package session

import (
	"fmt"
	"io"
	"kmud/database"
	"kmud/model"
	"kmud/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	// "log"
	// "os"
	"strings"
	"time"
)

type Session struct {
	conn   io.ReadWriter
	user   *database.User
	player *database.Character
	room   *database.Room

	prompt string

	userInputChannel chan string
	inputModeChannel chan userInputMode
	prompterChannel  chan utils.Prompter
	panicChannel     chan interface{}
	eventChannel     chan model.Event

	silentMode bool

	// These handlers encapsulate all of the functions that handle user
	// input into a single struct. This makes it so that we can use reflection
	// to lookup the function that should handle the user's input without fear
	// of them calling some function that we didn't intent to expose to them.
	commander commandHandler
	actioner  actionHandler

	replyId bson.ObjectId

	// logger *log.Logger
}

func NewSession(conn io.ReadWriter, user *database.User, player *database.Character) *Session {
	var session Session
	session.conn = conn
	session.user = user
	session.player = player
	session.room = model.GetRoom(player.GetRoomId())

	session.prompt = "%h/%H> "

	session.userInputChannel = make(chan string)
	session.inputModeChannel = make(chan userInputMode)
	session.prompterChannel = make(chan utils.Prompter)
	session.panicChannel = make(chan interface{})
	session.eventChannel = model.Register()

	session.silentMode = false
	session.commander.session = &session
	session.actioner.session = &session

	// file, err := os.OpenFile(player.GetName()+".log", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	// utils.PanicIfError(err)

	// session.logger = log.New(file, player.GetName()+" ", log.LstdFlags)

	model.Login(player)

	return &session
}

type userInputMode int

const (
	CleanUserInput userInputMode = iota
	RawUserInput   userInputMode = iota
)

func (session *Session) Exec() {
	defer model.Unregister(session.eventChannel)
	defer model.Logout(session.player)

	session.printLineColor(utils.ColorWhite, "Welcome, "+session.player.GetName())
	session.printRoom()

	// Main routine in charge of actually reading input from the connection object,
	// also has built in throttling to limit how fast we are allowed to process
	// commands from the user.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				session.panicChannel <- r
			}
		}()

		throttler := utils.NewThrottler(200 * time.Millisecond)

		for {
			mode := <-session.inputModeChannel
			prompter := <-session.prompterChannel
			input := ""

			switch mode {
			case CleanUserInput:
				input = utils.GetUserInputP(session.conn, prompter, session.user.GetColorMode())
			case RawUserInput:
				input = utils.GetRawUserInputP(session.conn, prompter, session.user.GetColorMode())
			default:
				panic("Unhandled case in switch statement (userInputMode)")
			}

			throttler.Sync()
			session.userInputChannel <- input
		}
	}()

	// Main loop
	for {
		input := session.getUserInputP(RawUserInput, session)
		if input == "" || input == "logout" || input == "quit" {
			return
		}

		if strings.HasPrefix(input, "/") {
			session.commander.handleCommand(utils.Argify(input[1:]))
		} else {
			session.actioner.handleAction(utils.Argify(input))
		}
	}
}

func (session *Session) printLineColor(color utils.Color, line string, a ...interface{}) {
	session.user.WriteLine(utils.Colorize(color, fmt.Sprintf(line, a...)))
}

func (session *Session) printLine(line string, a ...interface{}) {
	session.printLineColor(utils.ColorWhite, line, a...)
}

func (session *Session) printError(err string, a ...interface{}) {
	session.printLineColor(utils.ColorRed, err, a...)
}

func (session *Session) printRoom() {
	playerList := model.PlayersIn(session.room, session.player)
	npcList := model.NpcsIn(session.room)
	area := model.GetArea(session.room.GetAreaId())
	session.printLine(session.room.ToString(playerList, npcList,
		model.GetItems(session.room.GetItemIds()), area))
}

func (session *Session) clearLine() {
	utils.ClearLine(session.conn)
}

func (session *Session) asyncMessage(message string) {
	session.clearLine()
	session.printLine(message)
}

// Same behavior as menu.Exec(), except that it uses getUserInput
// which doesn't block the event loop while waiting for input
func (session *Session) execMenu(menu *utils.Menu) (string, bson.ObjectId) {
	choice := ""
	var data bson.ObjectId

	for {
		menu.Print(session.conn, session.user.GetColorMode())
		choice = session.getUserInputP(CleanUserInput, menu)
		if menu.HasAction(choice) || choice == "" {
			data = menu.GetData(choice)
			break
		}

		if choice != "?" {
			session.printError("Invalid selection")
		}
	}
	return choice, data
}

// getUserInput allows us to retrieve user input in a way that doesn't block the
// event loop by using channels and a separate Go routine to grab
// either the next user input or the next event.
func (session *Session) getUserInputP(inputMode userInputMode, prompter utils.Prompter) string {
	session.inputModeChannel <- inputMode
	session.prompterChannel <- prompter

	for {
		select {
		case input := <-session.userInputChannel:
			return input
		case event := <-session.eventChannel:
			if session.silentMode || !event.IsFor(session.player) {
				continue
			}

			if event.Type() == model.TellEventType {
				tellEvent := event.(model.TellEvent)
				session.replyId = tellEvent.From.GetId()
			} else if event.Type() == model.CombatEventType {
				combatEvent := event.(model.CombatEvent)

				if combatEvent.Defender == session.player {
					session.player.Hit(combatEvent.Damage)
					if session.player.GetHitPoints() <= 0 {
						session.asyncMessage(">> You're dead <<")
						model.StopFight(combatEvent.Defender)
						model.StopFight(combatEvent.Attacker)
					}
				}
			} else if event.Type() == model.TimerEventType {
				if !model.InCombat(session.player) {
					oldHps := session.player.GetHitPoints()
					session.player.Heal(5)
					newHps := session.player.GetHitPoints()

					if oldHps != newHps {
						session.clearLine()
						session.user.Write(prompter.GetPrompt())
					}
				}
			}

			message := event.ToString(session.player)
			if message != "" {
				session.asyncMessage(message)
				session.user.Write(prompter.GetPrompt())
			}

		case quitMessage := <-session.panicChannel:
			panic(quitMessage)
		}
	}
}

func (session *Session) getUserInput(inputMode userInputMode, prompt string) string {
	return session.getUserInputP(inputMode, utils.SimplePrompter(prompt))
}

func (session *Session) GetPrompt() string {
	prompt := session.prompt
	prompt = strings.Replace(prompt, "%h", strconv.Itoa(session.player.GetHitPoints()), -1)
	prompt = strings.Replace(prompt, "%H", strconv.Itoa(session.player.GetHealth()), -1)

	return utils.Colorize(utils.ColorWhite, prompt)
}

func (session *Session) currentZone() *database.Zone {
	return model.GetZone(session.room.GetZoneId())
}

// vim: nocindent
