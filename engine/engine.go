package engine

import (
	"time"

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

	eventListener := events.Register("engine")

	go func() {
		for {
			event := <-eventListener.Channel

			if event.Type() == events.CreateEventType {
				/*
				   createEvent := event.(model.CreateEvent)

				   go func() {
				       for {
				           npc := (<-npcChannel).(*database.Character)
				           manage(npc)
				       }
				   }()
				*/
			}
		}
	}()
}

func manage(npc types.NPC) {
	go func() {
		throttler := utils.NewThrottler(1 * time.Second)

		for {
			if npc.GetRoaming() {
				room := model.GetRoom(npc.GetRoomId())
				exits := room.GetExits()
				exitToTake := utils.Random(0, len(exits)-1)
				model.MoveCharacter(npc, exits[exitToTake])
			}

			throttler.Sync()
		}
	}()
}
