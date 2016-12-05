package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Store struct {
	Container `bson:",inline"`

	Name   string
	RoomId types.Id
}

func NewStore(name string, roomId types.Id) *Store {
	store := &Store{
		Name:   utils.FormatName(name),
		RoomId: roomId,
	}

	dbinit(store)
	return store
}

func (self *Store) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Store) SetName(name string) {
	self.writeLock(func() {
		self.Name = utils.FormatName(name)
	})
}
