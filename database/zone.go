package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Zone struct {
	DbObject `bson:",inline"`

	Name string
}

func NewZone(name string) *Zone {
	var zone Zone

	zone.Name = utils.FormatName(name)

	zone.initDbObject(&zone)
	return &zone
}

func (self *Zone) GetType() types.ObjectType {
	return types.ZoneType
}

func (self *Zone) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Zone) SetName(name string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if name != self.Name {
		self.Name = utils.FormatName(name)
		objectModified(self)
	}
}

// vim: nocindent
