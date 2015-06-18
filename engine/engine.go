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
	defer events.Unregister(npc)

	go func() {
		for {
			event := <-eventChannel
			switch event.Type() {
			case events.TickEventType:
				if npc.GetRoaming() {
					room := model.GetRoom(npc.GetRoomId())
					exits := room.GetExits()
					exitToTake := utils.Random(0, len(exits)-1)
					model.MoveCharacter(npc, exits[exitToTake])
				}
			}
		}
	}()
}
