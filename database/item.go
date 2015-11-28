package database

import "github.com/Cristofori/kmud/utils"

type Item struct {
	DbObject `bson:",inline"`

	Name  string
	Value int
}

func NewItem(name string) *Item {
	item := &Item{
		Name: utils.FormatName(name),
	}
	item.init(item)
	return item
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

func (self *Item) GetValue() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Value
}

// vim: nocindent
