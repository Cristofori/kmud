package events

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Cristofori/kmud/types"
)

var fightsMutex sync.RWMutex

var fights map[types.Character]types.Character // Maps the attacker to the defender

func StartFight(attacker types.Character, defender types.Character) {
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

func StopFight(attacker types.Character) {
	fightsMutex.Lock()
	defer fightsMutex.Unlock()

	defender := fights[attacker]

	if defender != nil {
		delete(fights, attacker)
		Broadcast(CombatStopEvent{Attacker: attacker, Defender: defender})
	}
}

func InCombat(character types.Character) bool {
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

func StartCombatLoop() {
	fights = map[types.Character]types.Character{}

	go func() {
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
	}()
}

// vim: nocindent
