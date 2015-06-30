package combat

import (
	"time"

	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

var _combatInterval = 3 * time.Second

var combatMessages chan interface{}
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
	combatMessages <- combatStart{Attacker: attacker, Defender: defender}
}

func StopFight(attacker types.Character) {
	combatMessages <- combatStop{Attacker: attacker}
}

func InCombat(character types.Character) bool {
	query := combatQuery{Character: character, Ret: make(chan bool)}
	combatMessages <- query
	return <-query.Ret
}

func StopCombatLoop() {
	defer func() { recover() }()
	close(combatMessages)
}

func StartCombatLoop() {
	if fights != nil {
		return
	}

	fights = map[types.Character]types.Character{}
	combatMessages = make(chan interface{}, 1)

	go func() {
		defer func() { recover() }()
		throttler := utils.NewThrottler(_combatInterval)
		for {
			throttler.Sync()
			combatMessages <- combatTick(true)
		}
	}()

	go func() {
		for message := range combatMessages {
		Switch:
			switch m := message.(type) {
			case combatTick:
				for a, d := range fights {
					if a.GetRoomId() == d.GetRoomId() {
						dmg := utils.Random(1, 10)
						d.Hit(dmg)

						events.Broadcast(events.CombatEvent{Attacker: a, Defender: d, Damage: dmg})

						if d.GetHitPoints() <= 0 {
							doCombatStop(a)
							doCombatStop(d)
							events.Broadcast(events.DeathEvent{Character: d})
						}

					} else {
						doCombatStop(a)
					}
				}
			case combatStart:
				oldDefender, found := fights[m.Attacker]

				if m.Defender == oldDefender {
					break
				}

				if found {
					doCombatStop(m.Attacker)
				}

				fights[m.Attacker] = m.Defender

				events.Broadcast(events.CombatStartEvent{Attacker: m.Attacker, Defender: m.Defender})
			case combatStop:
				doCombatStop(m.Attacker)
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
			default:
				panic("Unhandled combat message")
			}
		}
		fights = nil
	}()
}

func doCombatStop(attacker types.Character) {
	defender := fights[attacker]

	if defender != nil {
		delete(fights, attacker)
		events.Broadcast(events.CombatStopEvent{Attacker: attacker, Defender: defender})
	}
}

// vim: nocindent
