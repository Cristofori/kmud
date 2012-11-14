package game

import (
	"kmud/database"
	"kmud/utils"
)

type mapBuilder struct {
	width    int
	height   int
	data     [][]mapTile
	userRoom database.Room
}

type mapTile struct {
	char  rune
	color utils.Color
}

func (self *mapTile) toString(cm utils.ColorMode) string {
	if self.char == ' ' {
		return string(self.char)
	}

	return utils.Colorize(cm, self.color, string(self.char))
}

func newMapBuilder(width int, height int) mapBuilder {
	var builder mapBuilder

	// Double the size to account for extra space to draw exits
	width *= 2
	height *= 2

	builder.data = make([][]mapTile, height)

	for y := 0; y < height; y += 1 {
		builder.data[y] = make([]mapTile, width)
	}

	builder.width = width
	builder.height = height

	for y := 0; y < height; y += 1 {
		for x := 0; x < width; x += 1 {
			builder.data[y][x].char = ' '
		}
	}

	return builder
}

func (self *mapBuilder) setUserRoom(room database.Room) {
	self.userRoom = room
}

func (self *mapBuilder) addRoom(room database.Room, x int, y int) {
	addIfExists := func(dir database.ExitDirection, x int, y int) {
		if x < 0 || y < 0 {
			return
		}

		if room.HasExit(dir) {
			self.data[y][x].addExit(dir)
		}
	}

	if self.userRoom.Id == room.Id {
		self.data[y][x].char = 'O'
		self.data[y][x].color = utils.ColorRed
	} else {
		self.data[y][x].char = '#'
		self.data[y][x].color = utils.ColorWhite
	}

	addIfExists(database.DirectionNorth, x, y-1)
	addIfExists(database.DirectionNorthEast, x+1, y-1)
	addIfExists(database.DirectionEast, x+1, y)
	addIfExists(database.DirectionSouthEast, x+1, y+1)
	addIfExists(database.DirectionSouth, x, y+1)
	addIfExists(database.DirectionSouthWest, x-1, y+1)
	addIfExists(database.DirectionWest, x-1, y)
	addIfExists(database.DirectionNorthWest, x-1, y-1)
}

func (self *mapBuilder) toString(cm utils.ColorMode) string {
	var rows []string

	for y := 0; y < self.height; y += 1 {
		row := ""
		for x := 0; x < self.width; x += 1 {
			tile := self.data[y][x].toString(cm)
			row = row + tile
		}
		rows = append(rows, row)
	}
	rows = trim(rows)

	str := ""
	for _, row := range rows {
		str = str + row + "\n"
	}

	return str
}

func trim(rows []string) []string {
	rowEmpty := func(row string) bool {
		for _, char := range row {
			if char != ' ' {
				return false
			}
		}
		return true
	}

	// Trim from the top
	for _, row := range rows {
		if !rowEmpty(row) {
			break
		}

		rows = rows[1:]
	}

	// Trim from the bottom
	for i := len(rows) - 1; i >= 0; i -= 1 {
		row := rows[i]
		if !rowEmpty(row) {
			break
		}
		rows = rows[:len(rows)-1]
	}

	return rows
}

func (self *mapTile) addExit(dir database.ExitDirection) {
	combineChars := func(r1 rune, r2 rune, r3 rune) {
		if self.char == r1 {
			self.char = r2
		} else {
			self.char = r3
		}
	}

	self.color = utils.ColorBlue

	switch dir {
	case database.DirectionNorth:
		combineChars('v', '|', '^')
	case database.DirectionNorthEast:
		combineChars('\\', 'X', '/')
	case database.DirectionEast:
		combineChars('<', '-', '>')
	case database.DirectionSouthEast:
		combineChars('/', 'X', '\\')
	case database.DirectionSouth:
		combineChars('^', '|', 'v')
	case database.DirectionSouthWest:
		combineChars('\\', 'X', '/')
	case database.DirectionWest:
		combineChars('>', '-', '<')
	case database.DirectionNorthWest:
		combineChars('/', 'X', '\\')
	default:
		panic("Unexpected direction given to mapTile::addExit()")
	}
}

// vim: nocindent
