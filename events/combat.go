package events

import (
	"math/rand"
	"time"

	"github.com/Cristofori/kmud/types"
)

var _combatInterval = 3 * time.Second

var messageChannel chan interface{}
var fights map[types.Character]types.Character

type combatStart struct {
	Attacker types.Character
	Defender types.Character
}

type combatStop struct {
	Attacker types.Character
}

type combatQuery struct {
	Character types.Character
	Ret       chan bool
}

type combatTick bool

func StartFight(attacker types.Character, defender types.Character) {
	messageChannel <- combatStart{Attacker: attacker, Defender: defender}
}

func StopFight(attacker types.Character) {
	messageChannel <- combatStop{Attacker: attacker}
}

func InCombat(character types.Character) bool {
	query := combatQuery{Character: character, Ret: make(chan bool)}
	messageChannel <- query
	return <-query.Ret
}

func StopCombatLoop() {
	close(messageChannel)
}

func StartCombatLoop() {
	if fights != nil {
		return
	}

	fights = map[types.Character]types.Character{}
	messageChannel = make(chan interface{}, 1)

	go func() {
		defer func() { recover() }()
		for {
			time.Sleep(_combatInterval)
			messageChannel <- combatTick(true)
		}
	}()

	go func() {
		for message := range messageChannel {
		Switch:
			switch m := message.(type) {
			case combatTick:
				for a, d := range fights {
					if a.GetRoomId() == d.GetRoomId() {
						dmg := rand.Int()%10 + 1
						Broadcast(CombatEvent{Attacker: a, Defender: d, Damage: dmg})
					} else {
						StopFight(a)
					}
				}
			case combatStart:
				oldDefender, found := fights[m.Attacker]

				if m.Defender == oldDefender {
					break
				}

				if found {
					StopFight(m.Attacker)
				}

				fights[m.Attacker] = m.Defender

				Broadcast(CombatStartEvent{Attacker: m.Attacker, Defender: m.Defender})
			case combatStop:
				defender := fights[m.Attacker]

				if defender != nil {
					delete(fights, m.Attacker)
					Broadcast(CombatStopEvent{Attacker: m.Attacker, Defender: defender})
				}
			case combatQuery:
				_, found := fights[m.Character]

				if found {
					m.Ret <- true
				} else {
					for _, defender := range fights {
						if defender == m.Character {
							m.Ret <- true
							break Switch
						}
					}
					m.Ret <- false
				}
			}
		}
		fights = nil
	}()
}

// vim: nocindent
