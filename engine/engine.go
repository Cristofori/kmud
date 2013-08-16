package engine

import (
	"fmt"
	"kmud/database"
	"kmud/model"
    "kmud/utils"
    "time"
)

func Start() {
	for _, npc := range model.M.GetAllNpcs() {
		go manage(npc)
	}
}

func manage(npc *database.Character) {
    throttler := utils.NewThrottler(1 * time.Second)

	for {
        room := model.M.GetRoom(npc.GetRoomId())
        exits := room.GetExits()
        exitToTake := utils.Random(-1, len(exits) - 1)

        if exitToTake != -1 {
            fmt.Printf("Moving to exit %s\n", database.DirectionToString(exits[exitToTake]))
            model.MoveCharacter(npc, exits[exitToTake])
        } else {
            fmt.Println("Staying put")
        }

        throttler.Sync()
    }
}

