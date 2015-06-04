package database

import (
	"gopkg.in/mgo.v2/bson"
)

type Area struct {
	DbObject `bson:",inline"`

	Name   string
	ZoneId bson.ObjectId
}

type Areas []*Area

func NewArea(name string, zone bson.ObjectId) *Area {
	var area Area
	area.initDbObject()

	area.ZoneId = zone
	area.Name = name

	modified(&area)
	return &area
}

func (self *Area) GetType() objectType {
	return AreaType
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
		modified(self)
	}
}

func (self *Area) GetZoneId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.ZoneId
}
