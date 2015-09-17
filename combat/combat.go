package combat

import (
	"time"

	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

var combatInterval = 3 * time.Second

var combatMessages chan interface{}

var fights map[types.Character]types.Character
var skills map[types.Character]types.Skill

type combatStart struct {
	Attacker types.Character
	Defender types.Character
	Skill    types.Skill
}

type combatStop struct {
	Attacker types.Character
}

type combatQuery struct {
	Character types.Character
	Ret       chan bool
}

type combatTick bool

func StartFight(attacker types.Character, skill types.Skill, defender types.Character) {
	combatMessages <- combatStart{Attacker: attacker, Defender: defender, Skill: skill}
}

func StopFight(attacker types.Character) {
	combatMessages <- combatStop{Attacker: attacker}
}

func InCombat(character types.Character) bool {
	query := combatQuery{Character: character, Ret: make(chan bool)}
	combatMessages <- query
	return <-query.Ret
}

func init() {
	fights = map[types.Character]types.Character{}
	skills = map[types.Character]types.Skill{}

	combatMessages = make(chan interface{}, 1)

	go func() {
		defer func() { recover() }()
		throttler := utils.NewThrottler(combatInterval)
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
						var power int
						skill := skills[a]
						if skill == nil {
							power = utils.Random(1, 10)
						} else {
							power = skill.GetPower()
							variance := utils.Random(-skill.GetVariance(), skill.GetVariance())
							power += variance
						}

						d.Hit(power)
						events.Broadcast(events.CombatEvent{Attacker: a, Defender: d, Skill: skill, Power: power})

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
				skills[m.Attacker] = m.Skill

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
	}()
}

func doCombatStop(attacker types.Character) {
	defender := fights[attacker]

	if defender != nil {
		delete(fights, attacker)
		delete(skills, attacker)
		events.Broadcast(events.CombatStopEvent{Attacker: attacker, Defender: defender})
	}
}

// vim: nocindent
