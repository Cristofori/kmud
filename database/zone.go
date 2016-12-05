package database

import "github.com/Cristofori/kmud/utils"

type Zone struct {
	DbObject `bson:",inline"`

	Name string
}

func NewZone(name string) *Zone {
	zone := &Zone{
		Name: utils.FormatName(name),
	}

	dbinit(zone)
	return zone
}

func (self *Zone) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Name
}

func (self *Zone) SetName(name string) {
	self.writeLock(func() {
		self.Name = utils.FormatName(name)
	})
}
