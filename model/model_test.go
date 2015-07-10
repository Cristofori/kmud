package model

import (
	"testing"

	"github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/types"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
)

func Test(t *testing.T) { TestingT(t) }

type ModelSuite struct{}

var _ = Suite(&ModelSuite{})

var _db *mgo.Database

func (s *ModelSuite) SetUpSuite(c *C) {
	session, err := mgo.Dial("localhost")
	c.Assert(err, Equals, nil)

	if err != nil {
		return
	}

	dbName := "unit_model_test"

	_db = session.DB(dbName)

	session.DB(dbName).DropDatabase()
	Init(database.NewMongoSession(session), dbName)

	events.StartEvents()
}

func (s *ModelSuite) TearDownSuite(c *C) {
	_db.DropDatabase()
	datastore.ClearAll()
}

func (s *ModelSuite) TestUserFunctions(c *C) {
	name1 := "Test_name1"
	password1 := "test_password2"

	user1 := CreateUser(name1, password1)

	c.Assert(user1.GetName(), Equals, name1)
	c.Assert(user1.VerifyPassword(password1), Equals, true)

	user2 := GetOrCreateUser(name1, password1)
	c.Assert(user1, Equals, user2)

	name2 := "test_name2"
	password2 := "test_password2"
	user3 := GetOrCreateUser(name2, password2)
	c.Assert(user3, Not(Equals), user2)
	c.Assert(user3, Not(Equals), user1)

	userByName := GetUserByName(name1)
	c.Assert(userByName, Equals, user1)

	userByName = GetUserByName("foobar")
	c.Assert(userByName, Equals, nil)

	zone, _ := CreateZone("testZone")
	room, _ := CreateRoom(zone, types.Coordinate{X: 0, Y: 0, Z: 0})
	CreatePlayerCharacter("testPlayer", user1.GetId(), room)

	DeleteUser(user1.GetId())
	userByName = GetUserByName(name1)
	c.Assert(userByName, Equals, nil)
	c.Assert(GetUserCharacters(user1.GetId()), HasLen, 0)
}

func (s *ModelSuite) TestZoneFunctions(c *C) {
	name := "zone1"
	zone1, err1 := CreateZone(name)

	c.Assert(zone1, Not(Equals), nil)
	c.Assert(err1, Equals, nil)

	zoneByName := GetZoneByName(name)
	c.Assert(zoneByName, Equals, zone1)

	zone2, err2 := CreateZone("zone2")
	c.Assert(zone2, Not(Equals), nil)
	c.Assert(err2, Equals, nil)

	zone3, err3 := CreateZone("zone3")
	c.Assert(zone3, Not(Equals), nil)
	c.Assert(err3, Equals, nil)

	zoneById := GetZone(zone1.GetId())
	c.Assert(zoneById, Equals, zone1)

	_, err := CreateZone("zone3")
	c.Assert(err, Not(Equals), nil)
}

func (s *ModelSuite) TestRoomFunctions(c *C) {
	zone, err := CreateZone("zone")
	c.Assert(zone, Not(Equals), nil)
	c.Assert(err, Equals, nil)

	room1, err1 := CreateRoom(zone, types.Coordinate{X: 0, Y: 0, Z: 0})
	c.Assert(room1, Not(Equals), nil)
	c.Assert(err1, Equals, nil)

	badRoom, shouldError := CreateRoom(zone, types.Coordinate{X: 0, Y: 0, Z: 0})
	c.Assert(badRoom, Equals, nil)
	c.Assert(shouldError, Not(Equals), nil)

	room2, err2 := CreateRoom(zone, types.Coordinate{X: 0, Y: 1, Z: 0})
	c.Assert(room2, Not(Equals), nil)
	c.Assert(err2, Equals, nil)

	room1.SetExitEnabled(types.DirectionSouth, true)
	room2.SetExitEnabled(types.DirectionNorth, true)

	c.Assert(room2.HasExit(types.DirectionNorth), Equals, true)
	DeleteRoom(room1)
	c.Assert(room2.HasExit(types.DirectionNorth), Equals, false)
}

func (s *ModelSuite) TestRoomAndZoneFunctions(c *C) {
	// ZoneCorners
	// GetRoomsInZone
}

func (s *ModelSuite) TestCharFunctions(c *C) {
	//user := CreateUser("user1", "")
	//playerName1 := "player1"
	//player1 := CreatePlayer(name1, user
}
