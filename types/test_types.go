package types

import (
	"gopkg.in/mgo.v2/bson"
)

type MockIdentifiable struct {
	Id   bson.ObjectId
	Type ObjectType
}

func (self MockIdentifiable) GetId() bson.ObjectId {
	return self.Id
}

func (self MockIdentifiable) GetType() ObjectType {
	return self.Type
}

type MockNameable struct {
	Name string
}

func (self MockNameable) GetName() string {
	return self.Name
}

func (self *MockNameable) SetName(name string) {
	self.Name = name
}

type MockDestroyable struct {
}

func (MockDestroyable) Destroy() {
}

func (self MockDestroyable) IsDestroyed() bool {
	return false
}

type MockReadLocker struct {
}

func (*MockReadLocker) ReadLock() {
}

func (*MockReadLocker) ReadUnlock() {
}

type MockObject struct {
	MockIdentifiable
	MockReadLocker
	MockDestroyable
}

type MockZone struct {
	MockIdentifiable
}

func NewMockZone() *MockZone {
	return &MockZone{
		MockIdentifiable{Id: bson.NewObjectId(), Type: ZoneType},
	}
}

type MockRoom struct {
	MockIdentifiable
}

func NewMockRoom() *MockRoom {
	return &MockRoom{
		MockIdentifiable{Id: bson.NewObjectId(), Type: RoomType},
	}
}

type MockUser struct {
	MockIdentifiable
}

func NewMockUser() *MockUser {
	return &MockUser{
		MockIdentifiable{Id: bson.NewObjectId(), Type: UserType},
	}
}

type MockCharacter struct {
	MockObject
	MockNameable
}

func (*MockCharacter) AddCash(int) {
}

func (*MockCharacter) GetCash() int {
	return 0
}

func (*MockCharacter) GetHealth() int {
	return 0
}

func (*MockCharacter) GetHitPoints() int {
	return 0
}

func (*MockCharacter) Heal(int) {
}

func (*MockCharacter) Hit(int) {
}

func (*MockCharacter) AddItem(bson.ObjectId) {
}

func (*MockCharacter) RemoveItem(bson.ObjectId) {
}

func (*MockCharacter) GetItemIds() []bson.ObjectId {
	return []bson.ObjectId{}
}

type MockPC struct {
	MockCharacter
	RoomId bson.ObjectId
}

func NewMockPC() *MockPC {
	return &MockPC{
		MockCharacter: MockCharacter{
			MockObject: MockObject{
				MockIdentifiable: MockIdentifiable{Id: bson.NewObjectId(), Type: PcType},
			},
			MockNameable: MockNameable{Name: "Mock PC"},
		},
		RoomId: bson.NewObjectId(),
	}
}

func (self MockPC) GetRoomId() bson.ObjectId {
	return self.RoomId
}

func (self MockPC) IsOnline() bool {
	return true
}

func (self MockPC) SetRoomId(bson.ObjectId) {
}
