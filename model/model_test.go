package model

import (
	"kmud/database/dbtest"
	tu "kmud/testutils"
	"testing"
)

func Test_Init(t *testing.T) {
	Init(&dbtest.TestSession{})

	tu.Assert(M.users != nil, t, "Init() failed to initialize users")
	tu.Assert(M.chars != nil, t, "Init() failed to initialize chars")
	tu.Assert(M.rooms != nil, t, "Init() failed to initialize rooms")
	tu.Assert(M.zones != nil, t, "Init() failed to initialize zones")
	tu.Assert(M.items != nil, t, "Init() failed to initialize items")
}

func Test_UserFunctions(t *testing.T) {
	name1 := "test_name1"
	password1 := "test_password2"

	user1 := M.CreateUser(name1, password1)

	tu.Assert(user1.GetName() == name1, t, "User creation failed, bad name:", user1.GetName(), name1)
	tu.Assert(user1.VerifyPassword(password1), t, "User creation failed, bad password")

	user2 := M.GetOrCreateUser(name1, password1)
	tu.Assert(len(M.users) == 1, t, "GetOrCreateUser() shouldn't have created a new user")
	tu.Assert(user1 == user2, t, "GetOrCreateUser() should have returned the user we alreayd created")

	name2 := "test_name2"
	password2 := "test_password2"
	user3 := M.GetOrCreateUser(name2, password2)
	tu.Assert(len(M.users) == 2, t, "GetOrCreateUser() should have created a new user")
	tu.Assert(user3 != user2 && user3 != user1, t, "GetOrCreateUser() shouldn't have returned an already existing user")

	userList := M.GetUsers()
	tu.Assert(userList.Contains(user1), t, "GetUsers() didn't return user1")
	tu.Assert(userList.Contains(user2), t, "GetUsers() didn't return user2")
	tu.Assert(userList.Contains(user3), t, "GetUsers() didn't return user3")

	userByName := M.GetUserByName(name1)
	tu.Assert(userByName == user1, t, "GetUserByName() failed to find user1", name1)

	userByName = M.GetUserByName("foobar")
	tu.Assert(userByName == nil, t, "GetUserByName() should have returned nill")

	M.DeleteUser(user1)
	userByName = M.GetUserByName(name1)
	tu.Assert(userByName == nil, t, "DeleteUser() failed to delete user1")
	userList = M.GetUsers()
	tu.Assert(!userList.Contains(user1), t, "GetUsers() shouldn't have user1 in it anymore")

	// Cleanup
	M.DeleteUser(user2)
	M.DeleteUser(user3)
	tu.Assert(len(M.users) == 0, t, "There shouldn't be any users left")
}

func TestZoneFunctions(t *testing.T) {
	name := "zone1"
	zone1, _ := M.CreateZone(name)

	tu.Assert(zone1 != nil, t, "Zone creation failed")

	zoneByName := M.GetZoneByName(name)
	tu.Assert(zoneByName == zone1, t, "GetZoneByName() failed")

	zone2, _ := M.CreateZone("zone2")
	zone3, _ := M.CreateZone("zone3")

	tu.Assert(zone2 != nil, t, "Failed to create zone2")
	tu.Assert(zone3 != nil, t, "Failed to create zone3")

	zoneList := M.GetZones()
	tu.Assert(zoneList.Contains(zone1), t, "GetZones() didn't return zone1")
	tu.Assert(zoneList.Contains(zone2), t, "GetZones() didn't return zone2")
	tu.Assert(zoneList.Contains(zone3), t, "GetZones() didn't return zone3")

	zoneById := M.GetZone(zone1.GetId())
	tu.Assert(zoneById == zone1, t, "GetZoneById() failed")

	// Cleanup
}

func TestRoomFunctions(t *testing.T) {
	//name := "room"
}

func TestRoomAndZoneFunctions(t *testing.T) {
	// ZoneCorners
	// GetRoomsInZone
}

func TestCharFunctions(t *testing.T) {
	//user := M.CreateUser("user1", "")
	//playerName1 := "player1"
	//player1 := M.CreatePlayer(name1, user
}

// vim: nocindent
