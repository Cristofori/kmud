package testutils

import (
	"github.com/Cristofori/kmud/types"
	"gopkg.in/mgo.v2/bson"
)

type MockId string

func (self MockId) String() string {
	return string(self)
}

func (self MockId) Hex() string {
	return string(self)
}

type MockIdentifiable struct {
	Id types.Id
}

func (self MockIdentifiable) GetId() types.Id {
	return self.Id
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
		MockIdentifiable{Id: bson.NewObjectId()},
	}
}

type MockRoom struct {
	MockIdentifiable
}

func NewMockRoom() *MockRoom {
	return &MockRoom{
		MockIdentifiable{Id: bson.NewObjectId()},
	}
}

type MockUser struct {
	MockIdentifiable
}

func NewMockUser() *MockUser {
	return &MockUser{
		MockIdentifiable{Id: bson.NewObjectId()},
	}
}

type MockContainer struct {
}

func (*MockContainer) AddCash(int) {
}

func (*MockContainer) GetCash() int {
	return 0
}

func (*MockContainer) RemoveCash(int) {
}

func (*MockContainer) AddItem(types.Id) {
}

func (*MockContainer) RemoveItem(types.Id) bool {
	return true
}

func (*MockContainer) GetItems() []types.Id {
	return []types.Id{}
}

type MockCharacter struct {
	MockObject
	MockNameable
	MockContainer
}

func (*MockCharacter) GetHealth() int {
	return 1
}

func (*MockCharacter) SetHealth(int) {
}

func (*MockCharacter) GetHitPoints() int {
	return 1
}

func (*MockCharacter) Heal(int) {
}

func (*MockCharacter) Hit(int) {
}

func (*MockCharacter) SetHitPoints(int) {
}

type MockPC struct {
	MockCharacter
	RoomId types.Id
}

func NewMockPC() *MockPC {
	return &MockPC{
		MockCharacter: MockCharacter{
			MockObject: MockObject{
				MockIdentifiable: MockIdentifiable{Id: bson.NewObjectId()},
			},
			MockNameable: MockNameable{Name: "Mock PC"},
		},
		RoomId: bson.NewObjectId(),
	}
}

func (self MockPC) GetRoomId() types.Id {
	return self.RoomId
}

func (self MockPC) IsOnline() bool {
	return true
}

func (self MockPC) SetRoomId(types.Id) {
}

func (self MockPC) GetSkills() []types.Id {
	return []types.Id{}
}

func (self MockPC) AddSkill(types.Id) {
}
