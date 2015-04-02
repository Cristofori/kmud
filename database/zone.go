package database

import (
	"kmud/utils"
)

type Zone struct {
	DbObject `bson:",inline"`

	Name string
}

func NewZone(name string) *Zone {
	var zone Zone

	zone.Name = utils.FormatName(name)

	zone.initDbObject()
    objectModified(&zone)
	return &zone
}

func (self *Zone) GetType() objectType {
	return ZoneType
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

type Zones []*Zone

func (self Zones) Contains(z *Zone) bool {
	for _, zone := range self {
		if z == zone {
			return true
		}
	}

	return false
}

// vim: nocindent
