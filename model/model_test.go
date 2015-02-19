package model

import (
	"kmud/database"
	"kmud/database/dbtest"
	"kmud/testutils"
	tu "kmud/testutils"
	"testing"
	"time"
)

func _cleanup(t *testing.T) {
	for _, item := range _items {
		DeleteItem(item)
	}
	tu.Assert(len(_items) == 0, t, "Failed to cleanup all items")

	for _, char := range _chars {
		DeleteCharacter(char)
	}
	tu.Assert(len(_chars) == 0, t, "Failed to cleanup all characters")

	for _, user := range _users {
		DeleteUser(user)
	}
	tu.Assert(len(_users) == 0, t, "Failed to cleanup all users")

	for _, room := range _rooms {
		DeleteRoom(room)
	}
	tu.Assert(len(_rooms) == 0, t, "Failed to cleanup all rooms")

	for _, zone := range _zones {
		DeleteZone(zone)
	}
	tu.Assert(len(_zones) == 0, t, "Failed to cleanup all zones")
}

func Test_Init(t *testing.T) {
	Init(&dbtest.TestSession{})

	tu.Assert(_users != nil, t, "Init() failed to initialize users")
	tu.Assert(_chars != nil, t, "Init() failed to initialize chars")
	tu.Assert(_rooms != nil, t, "Init() failed to initialize rooms")
	tu.Assert(_zones != nil, t, "Init() failed to initialize zones")
	tu.Assert(_items != nil, t, "Init() failed to initialize items")
}

func Test_UserFunctions(t *testing.T) {
	name1 := "Test_name1"
	password1 := "test_password2"

	user1 := CreateUser(name1, password1)

	tu.Assert(user1.GetName() == name1, t, "User creation failed, bad name:", user1.GetName(), name1)
	tu.Assert(user1.VerifyPassword(password1), t, "User creation failed, bad password")

	user2 := GetOrCreateUser(name1, password1)
	tu.Assert(len(_users) == 1, t, "GetOrCreateUser() shouldn't have created a new user")
	tu.Assert(user1 == user2, t, "GetOrCreateUser() should have returned the user we alreayd created")

	name2 := "test_name2"
	password2 := "test_password2"
	user3 := GetOrCreateUser(name2, password2)
	tu.Assert(len(_users) == 2, t, "GetOrCreateUser() should have created a new user")
	tu.Assert(user3 != user2 && user3 != user1, t, "GetOrCreateUser() shouldn't have returned an already existing user")

	userList := GetUsers()
	tu.Assert(userList.Contains(user1), t, "GetUsers() didn't return user1")
	tu.Assert(userList.Contains(user2), t, "GetUsers() didn't return user2")
	tu.Assert(userList.Contains(user3), t, "GetUsers() didn't return user3")

	userByName := GetUserByName(name1)
	tu.Assert(userByName == user1, t, "GetUserByName() failed to find user1", name1)

	userByName = GetUserByName("foobar")
	tu.Assert(userByName == nil, t, "GetUserByName() should have returned nill")

	DeleteUser(user1)
	userByName = GetUserByName(name1)
	tu.Assert(userByName == nil, t, "DeleteUser() failed to delete user1")
	userList = GetUsers()
	tu.Assert(!userList.Contains(user1), t, "GetUsers() shouldn't have user1 in it anymore")

	zone, _ := CreateZone("testZone")
	room, _ := CreateRoom(zone, database.Coordinate{X: 0, Y: 0, Z: 0})
	CreatePlayer("testPlayer", user1, room)

	DeleteUser(user1)
	tu.Assert(len(_chars) == 0, t, "Deleting a user should have deleted its characters")

	_cleanup(t)
}

func Test_ZoneFunctions(t *testing.T) {
	name := "zone1"
	zone1, err1 := CreateZone(name)

	tu.Assert(zone1 != nil && err1 == nil, t, "Zone creation failed")

	zoneByName := GetZoneByName(name)
	tu.Assert(zoneByName == zone1, t, "GetZoneByName() failed")

	zone2, err2 := CreateZone("zone2")
	tu.Assert(zone2 != nil && err2 == nil, t, "Failed to create zone2")

	zone3, err3 := CreateZone("zone3")
	tu.Assert(zone3 != nil && err3 == nil, t, "Failed to create zone3")

	zoneList := GetZones()
	tu.Assert(zoneList.Contains(zone1), t, "GetZones() didn't return zone1")
	tu.Assert(zoneList.Contains(zone2), t, "GetZones() didn't return zone2")
	tu.Assert(zoneList.Contains(zone3), t, "GetZones() didn't return zone3")

	zoneById := GetZone(zone1.GetId())
	tu.Assert(zoneById == zone1, t, "GetZoneById() failed")

	_, err := CreateZone("zone3")
	tu.Assert(err != nil, t, "Creating zone with duplicate name should have failed")

	_cleanup(t)
}

func Test_RoomFunctions(t *testing.T) {
	zone, err := CreateZone("zone")
	tu.Assert(zone != nil && err == nil, t, "Zone creation failed")

	room1, err1 := CreateRoom(zone, database.Coordinate{X: 0, Y: 0, Z: 0})
	tu.Assert(room1 != nil && err1 == nil, t, "Room creation failed")

	badRoom, shouldError := CreateRoom(zone, database.Coordinate{X: 0, Y: 0, Z: 0})
	tu.Assert(badRoom == nil && shouldError != nil, t, "Creating two rooms at the same location should have failed")

	room2, err2 := CreateRoom(zone, database.Coordinate{X: 0, Y: 1, Z: 0})
	tu.Assert(room2 != nil && err2 == nil, t, "Second room creation failed")

	room1.SetExitEnabled(database.DirectionSouth, true)
	room2.SetExitEnabled(database.DirectionNorth, true)

	tu.Assert(room2.HasExit(database.DirectionNorth), t, "Call to room.SetExitEnabled failed")
	DeleteRoom(room1)
	tu.Assert(!room2.HasExit(database.DirectionNorth), t, "Deleting room1 should have removed corresponding exit from room2")

	_cleanup(t)
}

func Test_RoomAndZoneFunctions(t *testing.T) {
	// ZoneCorners
	// GetRoomsInZone
}

func Test_CharFunctions(t *testing.T) {
	//user := CreateUser("user1", "")
	//playerName1 := "player1"
	//player1 := CreatePlayer(name1, user
}

func Test_EventLoop(t *testing.T) {
	zone, _ := CreateZone("zone")
	room, _ := CreateRoom(zone, database.Coordinate{X: 0, Y: 0, Z: 0})
	user := CreateUser("user", "password")
	char := CreatePlayer("char", user, room)

	eventChannel := Register()

	message := "hey how are yah"
	queueEvent(TellEvent{char, char, message})

	timeout := testutils.Timeout(3 * time.Second)

	select {
	case event := <-eventChannel:
		tu.Assert(event.Type() == TellEventType, t, "Didn't get a Tell event back")
		tellEvent := event.(TellEvent)
		tu.Assert(tellEvent.Message == message, t, "Didn't get the right message back:", tellEvent.Message, message)
	case <-timeout:
		tu.Assert(false, t, "Timed out waiting for tell event")
	}

	select {
	case event := <-eventChannel:
		tu.Assert(event.Type() == TimerEventType, t, "Expected to get a timer event")
	case <-timeout:
		tu.Assert(false, t, "Timed out waiting for timer event")
	}

	_cleanup(t)
}

func Test_CombatLoop(t *testing.T) {
	zone, _ := CreateZone("zone")
	room, _ := CreateRoom(zone, database.Coordinate{X: 0, Y: 0, Z: 0})
	user := CreateUser("user", "password")

	char1 := CreatePlayer("char1", user, room)
	char2 := CreatePlayer("char2", user, room)

	eventChannel1 := Register()
	// eventChannel2 := Register(char2)

	StartFight(char1, char2)
	// StartFight(char2, char1)

	verifyEvents := func(eventChannel chan Event) {
		timeout := testutils.Timeout(3 * time.Second)
		expectedTypes := make(map[EventType]bool)
		expectedTypes[CombatEventType] = true
		expectedTypes[CombatStartEventType] = true

		for {
			select {
			case event := <-eventChannel1:
				if event.Type() != TimerEventType {
					tu.Assert(expectedTypes[event.Type()] == true, t, "Unexpected event type:", event.Type())
					delete(expectedTypes, event.Type())
				}
			case <-timeout:
				tu.Assert(false, t, "Timed out waiting for combat event")
			}

			if len(expectedTypes) == 0 {
				break
			}
		}
	}

	verifyEvents(eventChannel1)
	// verifyEvents(eventChannel2)
}

// vim: nocindent
