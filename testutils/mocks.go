package testutils

import (
	"github.com/Cristofori/kmud/types"
	"gopkg.in/mgo.v2/bson"
)

type MockIdentifiable struct {
	Id   bson.ObjectId
	Type types.ObjectType
}

func (self MockIdentifiable) GetId() bson.ObjectId {
	return self.Id
}

func (self MockIdentifiable) GetType() types.ObjectType {
	return self.Type
}

type MockNameable struct {
	Name string
}

func (self MockNameable) GetName() string {
	return self.Name
}

type MockZone struct {
	MockIdentifiable
}

func Zone() MockZone {
	return MockZone{
		MockIdentifiable{Id: bson.NewObjectId(), Type: types.ZoneType},
	}
}

type MockRoom struct {
	MockIdentifiable
}

func Room() MockRoom {
	return MockRoom{
		MockIdentifiable{Id: bson.NewObjectId(), Type: types.RoomType},
	}
}

type MockUser struct {
	MockIdentifiable
}

func User() MockUser {
	return MockUser{
		MockIdentifiable{Id: bson.NewObjectId(), Type: types.UserType},
	}
}

type MockPlayerCharacter struct {
	MockIdentifiable
	MockNameable
	RoomId bson.ObjectId
}

func PlayerCharacter() MockPlayerCharacter {
	return MockPlayerCharacter{
		MockIdentifiable: MockIdentifiable{Id: bson.NewObjectId(), Type: types.PcType},
		MockNameable:     MockNameable{Name: "Mock PC"},
		RoomId:           bson.NewObjectId(),
	}
}

func (self MockPlayerCharacter) GetRoomId() bson.ObjectId {
	return self.RoomId
}

func (self MockPlayerCharacter) IsOnline() bool {
	return true
}

func (self MockPlayerCharacter) SetRoomId(bson.ObjectId) {
}
