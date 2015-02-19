package engine

import (
	"kmud/database"
	"kmud/model"
	"kmud/utils"
	"time"
)

const (
	RoamingProperty = "roaming"
)

func Start() {
	for _, npc := range model.GetAllNpcs() {
		manage(npc)
	}

	eventChannel := model.Register()

	event := <-eventChannel

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

func manage(npc *database.Character) {
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
