package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Container struct {
	DbObject  `bson:",inline"`
	Inventory utils.Set
	Cash      int
}

func (self *Container) AddItem(id types.Id) {
	if !self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		if self.Inventory == nil {
			self.Inventory = utils.Set{}
		}

		self.Inventory.Insert(id.Hex())
	}
	self.modified()
}

func (self *Container) HasItem(id types.Id) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Inventory.Contains(id.Hex())
}

func (self *Container) RemoveItem(id types.Id) bool {
	if self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		delete(self.Inventory, id.Hex())
		self.modified()
		return true
	}
	return false
}

func (self *Container) GetItems() []types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return idSetToList(self.Inventory)
}

func (self *Container) SetCash(cash int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if cash != self.Cash {
		self.Cash = cash
		self.modified()
	}
}

func (self *Container) GetCash() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Cash
}

func (self *Container) AddCash(amount int) {
	self.SetCash(self.GetCash() + amount)
}

func (self *Container) RemoveCash(amount int) {
	self.SetCash(self.GetCash() - amount)
}
