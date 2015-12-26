package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Area struct {
	DbObject `bson:",inline"`

	Name   string
	ZoneId types.Id
}

func NewArea(name string, zone types.Id) *Area {
	area := &Area{
		ZoneId: zone,
		Name:   utils.FormatName(name),
	}

	dbinit(area)
	return area
}

func (self *Area) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Area) SetName(name string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if name != self.Name {
		self.Name = name
		self.modified()
	}
}

func (self *Area) GetZoneId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.ZoneId
}
