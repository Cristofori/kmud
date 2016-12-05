package database

import "github.com/Cristofori/kmud/types"

type Exit struct {
	Locked bool
}

type Room struct {
	Container `bson:",inline"`

	ZoneId      types.Id
	AreaId      types.Id `bson:",omitempty"`
	Title       string
	Description string
	Links       map[string]types.Id
	Location    types.Coordinate

	Exits map[types.Direction]*Exit
}

func NewRoom(zoneId types.Id, location types.Coordinate) *Room {
	room := &Room{
		Title: "The Void",
		Description: "You are floating in the blackness of space. Complete darkness surrounds " +
			"you in all directions. There is no escape, there is no hope, just the emptiness. " +
			"You are likely to be eaten by a grue.",
		Location: location,
		ZoneId:   zoneId,
	}

	dbinit(room)
	return room
}

func (self *Room) HasExit(dir types.Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	_, found := self.Exits[dir]
	return found
}

func (self *Room) SetExitEnabled(dir types.Direction, enabled bool) {
	self.writeLock(func() {
		if self.Exits == nil {
			self.Exits = map[types.Direction]*Exit{}
		}
		if enabled {
			self.Exits[dir] = &Exit{}
		} else {
			delete(self.Exits, dir)
		}
	})
}

func (self *Room) SetLink(name string, roomId types.Id) {
	self.writeLock(func() {
		if self.Links == nil {
			self.Links = map[string]types.Id{}
		}
		self.Links[name] = roomId
	})
}

func (self *Room) RemoveLink(name string) {
	self.writeLock(func() {
		delete(self.Links, name)
	})
}

func (self *Room) GetLinks() map[string]types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Links
}

func (self *Room) LinkNames() []string {
	names := make([]string, len(self.GetLinks()))

	i := 0
	for name := range self.Links {
		names[i] = name
		i++
	}
	return names
}

func (self *Room) SetTitle(title string) {
	self.writeLock(func() {
		self.Title = title
	})
}

func (self *Room) GetTitle() string {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Title
}

func (self *Room) SetDescription(description string) {
	self.writeLock(func() {
		self.Description = description
	})
}

func (self *Room) GetDescription() string {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Description
}

func (self *Room) SetLocation(location types.Coordinate) {
	self.writeLock(func() {
		self.Location = location
	})
}

func (self *Room) GetLocation() types.Coordinate {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Location
}

func (self *Room) SetZoneId(zoneId types.Id) {
	self.writeLock(func() {
		self.ZoneId = zoneId
	})
}

func (self *Room) GetZoneId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.ZoneId
}

func (self *Room) SetAreaId(areaId types.Id) {
	self.writeLock(func() {
		self.AreaId = areaId
	})
}

func (self *Room) GetAreaId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.AreaId
}

func (self *Room) NextLocation(direction types.Direction) types.Coordinate {
	loc := self.GetLocation()
	return loc.Next(direction)
}

func (self *Room) GetExits() []types.Direction {
	self.ReadLock()
	defer self.ReadUnlock()

	exits := make([]types.Direction, len(self.Exits))

	i := 0
	for dir := range self.Exits {
		exits[i] = dir
		i++
	}

	return exits
}

func (self *Room) SetLocked(dir types.Direction, locked bool) {
	self.writeLock(func() {
		if self.HasExit(dir) {
			self.Exits[dir].Locked = locked
		}
	})
}

func (self *Room) IsLocked(dir types.Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	if self.HasExit(dir) {
		return self.Exits[dir].Locked
	}

	return false
}
