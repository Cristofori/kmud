package engine

import (
	"kmud/database"
	"kmud/model"
	"kmud/utils"
	"time"
)

func Start() {
	for _, npc := range model.M.GetAllNpcs() {
		manage(npc)
	}

	npcChannel := model.Watch(model.NewNpcUpdate)

	go func() {
		for {
			npc := (<-npcChannel).(*database.Character)
			manage(npc)
		}
	}()
}

func manage(npc *database.Character) {
	go func() {
		throttler := utils.NewThrottler(1 * time.Second)

		for {
			room := model.M.GetRoom(npc.GetRoomId())
			exits := room.GetExits()
			exitToTake := utils.Random(0, len(exits)-1)
			model.MoveCharacter(npc, exits[exitToTake])

			throttler.Sync()
		}
	}()
}
