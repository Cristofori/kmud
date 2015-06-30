package engine

import (
	"github.com/Cristofori/kmud/combat"
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"time"
)

const (
	RoamingProperty = "roaming"
)

func Start() {
	for _, npc := range model.GetNpcs() {
		manageNpc(npc)
	}

	for _, spawner := range model.GetSpawners() {
		manageSpawner(spawner)
	}
}

func manageNpc(npc types.NPC) {
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
					combat.StartFight(npc, e.Attacker)
				}
			case events.CombatStopEvent:
				if npc == e.Defender {
					combat.StopFight(npc)
				}
			case events.DeathEvent:
				if npc == e.Character {
					model.DeleteCharacter(npc)
					return
				}
			}
		}
	}()
}

func manageSpawner(spawner types.Spawner) {
	throttler := utils.NewThrottler(5 * time.Second)
	go func() {
		for {
			rooms := model.GetAreaRooms(spawner.GetAreaId())

			if len(rooms) > 0 {
				npcs := model.GetSpawnerNpcs(spawner.GetId())
				diff := spawner.GetCount() - len(npcs)

				for diff > 0 && len(rooms) > 0 {
					room := rooms[utils.Random(0, len(rooms)-1)]
					npc := model.CreateNpc(spawner.GetName(), room.GetId(), spawner.GetId())
					npc.SetHealth(spawner.GetHealth())
					manageNpc(npc)
					diff--
				}
			}

			throttler.Sync()
		}
	}()
}
