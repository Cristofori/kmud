package database

import (
	"github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/utils"
)

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

func (self *Item) GetType() datastore.ObjectType {
	return ItemType
}

func (self *Item) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func ItemNames(items []*Item) []string {
	names := make([]string, len(items))

	for i, item := range items {
		names[i] = item.GetName()
	}

	return names
}

// vim: nocindent
