package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
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

	area.ZoneId = zone
	area.Name = utils.FormatName(name)

	area.initDbObject(&area)

	return &area
}

func (self *Area) GetType() types.ObjectType {
	return types.AreaType
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
