package engine

import (
	"fmt"
	"kmud/database"
	"kmud/model"
)

func Start() {
	for _, npc := range model.M.GetAllNpcs() {
		go manage(npc)
	}
}

func manage(npc *database.Character) {
	fmt.Printf("Managing: %v\n", npc.GetName())
}

