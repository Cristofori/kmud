package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Store struct {
	DbObject  `bson:",inline"`
	Container `bson:",inline"`

	Name   string
	RoomId types.Id
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
	self.Container.AddItem(id)
	self.modified()
}

func (self *Store) RemoveItem(id types.Id) bool {
	self.modified()
	return self.Container.RemoveItem(id)
}

func (self *Store) SetCash(cash int) {
	self.Container.SetCash(cash)
	self.modified()
}

func (self *Store) AddCash(amount int) {
	self.SetCash(self.GetCash() + amount)
}

func (self *Store) RemoveCash(amount int) {
	self.SetCash(self.GetCash() - amount)
}
