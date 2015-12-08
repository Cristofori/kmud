package engine

import (
	"math"

	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

func costEstimate(start, goal types.Room) int {
	c1 := start.GetLocation()
	c2 := goal.GetLocation()

	return utils.Abs(c1.X-c2.X) + utils.Abs(c1.Y-c2.Y) + utils.Abs(c1.Z-c2.Z)
}

type roomSet map[types.Room]bool

func lowest(rooms roomSet, scores map[types.Room]int) types.Room {
	var lowest types.Room
	lowestValue := math.MaxInt32

	for room := range rooms {
		value, found := scores[room]
		if found && value < lowestValue {
			lowest = room
			lowestValue = value
		}
	}

	return lowest
}

func reconstruct(cameFrom map[types.Room]types.Room, current types.Room) []types.Room {
	path := []types.Room{current}

	for {
		found := false
		current, found = cameFrom[current]
		if !found {
			break
		}

		path = append([]types.Room{current}, path...)
	}

	return path

}

// A* pathfinding algorithm, adapted from wikipedia's pseudocode
// TODO - Find out what happens if the rooms aren't in the same zone
func FindPath(start, goal types.Room) []types.Room {
	/*
		t1 := time.Now()
		defer func() {
			fmt.Printf("Took %v to find a path from %v to %v in %s\n", time.Since(t1), start.GetLocation(), goal.GetLocation(), model.GetZone(start.GetZoneId()).GetName())
		}()
	*/

	evaluated := roomSet{}

	unevaluated := roomSet{}
	unevaluated[start] = true

	cameFrom := map[types.Room]types.Room{}

	g_score := map[types.Room]int{}
	g_score[start] = 0

	f_score := map[types.Room]int{}
	f_score[start] = costEstimate(start, goal)

	for {
		if len(unevaluated) == 0 {
			break
		}

		current := lowest(unevaluated, f_score)
		if current == goal {
			return reconstruct(cameFrom, goal)
		}

		delete(unevaluated, current)
		evaluated[current] = true

		neighbors := model.GetNeighbors(current)

		for _, neighbor := range neighbors {
			_, found := evaluated[neighbor]
			if found {
				continue
			}

			_, found = g_score[neighbor]
			if !found {
				g_score[neighbor] = math.MaxInt32
			}

			tentative_g_score := g_score[current] + 1 // TODO - Update this when there is a travel penalty between rooms

			_, found = unevaluated[neighbor]

			if !found {
				unevaluated[neighbor] = true
			} else if tentative_g_score >= g_score[neighbor] {
				continue
			}

			cameFrom[neighbor] = current
			g_score[neighbor] = tentative_g_score
			f_score[neighbor] = g_score[neighbor] + costEstimate(neighbor, goal)
		}
	}

	return []types.Room{}
}
