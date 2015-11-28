package database

import (
	"sync"

	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Locker struct {
	mutex sync.RWMutex
}

type Container struct {
	Locker    `bson:",omitempty"`
	Inventory utils.Set
	Cash      int
}

func (self *Locker) ReadLock() {
	self.mutex.RLock()
}

func (self *Locker) ReadUnlock() {
	self.mutex.RUnlock()
}

func (self *Locker) WriteLock() {
	self.mutex.Lock()
}

func (self *Locker) WriteUnlock() {
	self.mutex.Unlock()
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
	}
}

func (self *Container) GetCash() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Cash
}
