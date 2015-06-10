package database

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/utils"
)

type Area struct {
	DbObject `bson:",inline"`

	Name   string
	ZoneId bson.ObjectId
}

type Areas []*Area

func NewArea(name string, zone bson.ObjectId) *Area {
	var area Area

	area.ZoneId = zone
	area.Name = utils.FormatName(name)

	area.initDbObject(&area)

	return &area
}

func (self *Area) GetType() datastore.ObjectType {
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
		objectModified(self)
	}
}

func (self *Area) GetZoneId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.ZoneId
}
