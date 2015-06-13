package model

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Cristofori/kmud/database"
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

	Broadcast(CombatStartEvent{Attacker: attacker, Defender: defender})
}

func StopFight(attacker *database.Character) {
	fightsMutex.Lock()
	defer fightsMutex.Unlock()

	defender := fights[attacker]

	if defender != nil {
		delete(fights, attacker)
		Broadcast(CombatStopEvent{Attacker: attacker, Defender: defender})
	}
}

func InCombat(character *database.Character) bool {
	_, found := fights[character]

	if found {
		return true
	}

	for _, defender := range fights {
		if defender == character {
			return true
		}
	}
	return false
}

func combatLoop() {
	for {
		time.Sleep(3 * time.Second)

		fightsMutex.RLock()
		for a, d := range fights {
			if a.GetRoomId() == d.GetRoomId() {
				dmg := rand.Int()%10 + 1
				Broadcast(CombatEvent{Attacker: a, Defender: d, Damage: dmg})
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
