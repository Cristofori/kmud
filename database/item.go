package database

import "github.com/Cristofori/kmud/utils"

type Item struct {
	DbObject `bson:",inline"`

	Name string
}

func NewItem(name string) *Item {
	var item Item
	item.Name = utils.FormatName(name)
	item.initDbObject(&item)
	return &item
}

func (self *Item) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Item) SetName(name string) {
	if name != self.GetName() {
		self.WriteLock()
		self.Name = utils.FormatName(name)
		self.WriteUnlock()
		self.modified()
	}
}

// vim: nocindent
