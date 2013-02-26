package model

import (
	"kmud/database"
	"sync"
	"time"
)

var fightsMutex sync.RWMutex

var fights map[*database.Character]*database.Character // Maps the attacker to the defender

func StartFight(attacker *database.Character, defender *database.Character) {
	fightsMutex.Lock()
	defer fightsMutex.Unlock()

	oldDefender, found := fights[attacker]

	if defender == oldDefender {
		return
	}

	if found {
		fightsMutex.Unlock()
		StopFight(attacker)
		fightsMutex.Lock()
	}

	fights[attacker] = defender

	queueEvent(CombatStartEvent{Attacker: attacker, Defender: defender})
}

func StopFight(attacker *database.Character) {
	fightsMutex.Lock()
	defer fightsMutex.Unlock()

	defender := fights[attacker]

	if defender != nil {
		delete(fights, attacker)
		queueEvent(CombatStopEvent{Attacker: attacker, Defender: defender})
	}
}

func combatLoop() {
	for {
		time.Sleep(3 * time.Second)

		fightsMutex.RLock()
		for a, d := range fights {
			if a.GetRoomId() == d.GetRoomId() {
				queueEvent(CombatEvent{Attacker: a, Defender: d, Damage: 10})
			} else {
				fightsMutex.RUnlock()
				StopFight(a)
				fightsMutex.RLock()
			}
		}
		fightsMutex.RUnlock()
	}
}

// vim: nocindent
