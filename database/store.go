package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Store struct {
	DbObject `bson:",inline"`

	Name      string
	Inventory utils.Set
	RoomId    types.Id
	Cash      int
}

func NewStore(name string, roomId types.Id) *Store {
	store := &Store{
		Name:   utils.FormatName(name),
		RoomId: roomId,
	}

	store.init(store)
	return store
}

func (self *Store) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Store) SetName(name string) {
	if name != self.GetName() {
		self.WriteLock()
		self.Name = utils.FormatName(name)
		self.WriteUnlock()
		self.modified()
	}
}

func (self *Store) AddItem(id types.Id) {
	if !self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		if self.Inventory == nil {
			self.Inventory = utils.Set{}
		}

		self.Inventory.Insert(id.Hex())
		self.modified()
	}
}

func (self *Store) HasItem(id types.Id) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Inventory.Contains(id.Hex())
}

func (self *Store) RemoveItem(id types.Id) bool {
	if self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		delete(self.Inventory, id.Hex())
		self.modified()
		return true
	}
	return false
}

func (self *Store) GetItems() []types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return idSetToList(self.Inventory)
}

func (self *Store) SetCash(cash int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if cash != self.Cash {
		self.Cash = cash
		self.modified()
	}
}

func (self *Store) AddCash(amount int) {
	self.SetCash(self.GetCash() + amount)
}

func (self *Store) RemoveCash(amount int) {
	self.SetCash(self.GetCash() - amount)
}

func (self *Store) GetCash() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Cash
}
