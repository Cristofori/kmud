package game

import (
	"kmud/database"
	"kmud/utils"
)

type mapBuilder struct {
	width    int
	height   int
	depth    int
	data     [][][]mapTile
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

func newMapBuilder(width int, height int, depth int) mapBuilder {
	var builder mapBuilder

	// Double the X/Y axis to account for extra space to draw exits
	width *= 2
	height *= 2

	builder.data = make([][][]mapTile, depth)

	for z := 0; z < depth; z++ {
		builder.data[z] = make([][]mapTile, height)
		for y := 0; y < height; y++ {
			builder.data[z][y] = make([]mapTile, width)
		}
	}

	builder.width = width
	builder.height = height
	builder.depth = depth

	for z := 0; z < depth; z++ {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				builder.data[z][y][x].char = ' '
			}
		}
	}

	return builder
}

func (self *mapBuilder) setUserRoom(room database.Room) {
	self.userRoom = room
}

func (self *mapBuilder) addRoom(room database.Room, x int, y int, z int) {
	addIfExists := func(dir database.ExitDirection, x int, y int) {
		if x < 0 || y < 0 {
			return
		}

		if room.HasExit(dir) {
			self.data[z][y][x].addExit(dir)
		}
	}

	if self.userRoom.Id == room.Id {
		self.data[z][y][x].char = 'O'
		self.data[z][y][x].color = utils.ColorRed
	} else {
		self.data[z][y][x].color = utils.ColorMagenta
		if room.HasExit(database.DirectionUp) && room.HasExit(database.DirectionDown) {
			self.data[z][y][x].char = '+'
		} else if room.HasExit(database.DirectionUp) {
			self.data[z][y][x].char = 'U'
		} else if room.HasExit(database.DirectionDown) {
			self.data[z][y][x].char = 'D'
		} else {
			self.data[z][y][x].char = '#'
			self.data[z][y][x].color = utils.ColorWhite
		}
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
	str := ""

	for z := 0; z < self.depth; z++ {
		var rows []string
		for y := 0; y < self.height; y++ {
			row := ""
			for x := 0; x < self.width; x++ {
				tile := self.data[z][y][x].toString(cm)
				row = row + tile
			}
			rows = append(rows, row)
		}

		if self.depth > 1 {
			divider := "================================================================================\n"
			rows = append(rows, divider)
		}

		for _, row := range rows {
			str = str + row + "\n"
		}
	}

	return str
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
		combineChars('|', '|', '|')
	case database.DirectionNorthEast:
		combineChars('\\', 'X', '/')
	case database.DirectionEast:
		combineChars('-', '-', '-')
	case database.DirectionSouthEast:
		combineChars('/', 'X', '\\')
	case database.DirectionSouth:
		combineChars('|', '|', '|')
	case database.DirectionSouthWest:
		combineChars('\\', 'X', '/')
	case database.DirectionWest:
		combineChars('-', '-', '-')
	case database.DirectionNorthWest:
		combineChars('/', 'X', '\\')
	default:
		panic("Unexpected direction given to mapTile::addExit()")
	}
}

// vim: nocindent
