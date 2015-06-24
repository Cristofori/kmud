package engine

import (
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

const (
	RoamingProperty = "roaming"
)

func Start() {
	for _, npc := range model.GetNpcs() {
		manage(npc)
	}
}

func manage(npc types.NPC) {
	eventChannel := events.Register(npc)

	go func() {
		defer events.Unregister(npc)

		for {
			event := <-eventChannel
			switch e := event.(type) {
			case events.TickEvent:
				if npc.GetRoaming() {
					room := model.GetRoom(npc.GetRoomId())
					exits := room.GetExits()
					exitToTake := utils.Random(0, len(exits)-1)
					model.MoveCharacter(npc, exits[exitToTake])
				}
			case events.CombatStartEvent:
				if npc == e.Defender {
					events.StartFight(npc, e.Attacker)
				}
			case events.CombatStopEvent:
				if npc == e.Defender {
					events.StopFight(npc)
				}
			}
		}
	}()
}
