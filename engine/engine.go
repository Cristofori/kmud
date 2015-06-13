package engine

import (
	"time"

	"github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/utils"
)

const (
	RoamingProperty = "roaming"
)

func Start() {
	for _, npc := range model.GetNpcs() {
		manage(npc)
	}

	eventListener := model.Register("engine")

	go func() {
		for {
			event := <-eventListener.Channel

			if event.Type() == model.CreateEventType {
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

func manage(npc *database.NonPlayerChar) {
	go func() {
		throttler := utils.NewThrottler(1 * time.Second)

		for {
			if npc.GetRoaming() {
				room := model.GetRoom(npc.GetRoomId())
				exits := room.GetExits()
				exitToTake := utils.Random(0, len(exits)-1)
				model.MoveCharacter(&npc.Character, exits[exitToTake])
			}

			throttler.Sync()
		}
	}()
}
