package dbtest

import (
	"kmud/database"
	"kmud/testutils"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

// import "fmt"

type TestSession struct {
}

func (ms TestSession) DB(dbName string) database.Database {
	return &TestDatabase{}
}

type TestDatabase struct {
}

func (md TestDatabase) C(collectionName string) database.Collection {
	return &TestCollection{}
}

type TestCollection struct {
}

func (mc TestCollection) Find(selector interface{}) database.Query {
	return &TestQuery{}
}

func (mc TestCollection) RemoveId(id interface{}) error {
	return nil
}

func (mc TestCollection) Remove(selector interface{}) error {
	return nil
}

func (mc TestCollection) DropCollection() error {
	return nil
}

func (mc TestCollection) UpdateId(id interface{}, change interface{}) error {
	return nil
}

func (mc TestCollection) UpsertId(id interface{}, change interface{}) error {
	return nil
}

type TestQuery struct {
}

func (mq TestQuery) Count() (int, error) {
	return 0, nil
}

func (mq TestQuery) One(result interface{}) error {
	return nil
}

func (mq TestQuery) Iter() database.Iterator {
	return &TestIterator{}
}

type TestIterator struct {
}

func (mi TestIterator) All(result interface{}) error {
	return nil
}

func Test_ThreadSafety(t *testing.T) {
	runtime.GOMAXPROCS(2)
	database.Init(&TestSession{})

	char := database.NewCharacter("test", "", "")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < 100; i++ {
			char.SetName(strconv.Itoa(i))
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 100; i++ {
		}
		wg.Done()
	}()

	wg.Wait()
}

func Test_User(t *testing.T) {
	user := database.NewUser("testuser", "")

	if user.Online() {
		t.Errorf("Newly created user shouldn't be online")
	}

	user.SetOnline(true)

	testutils.Assert(user.Online(), t, "Call to SetOnline(true) failed")
	testutils.Assert(user.GetColorMode() == utils.ColorModeNone, t, "Newly created user should have a color mode of None")

	user.SetColorMode(utils.ColorModeLight)

	testutils.Assert(user.GetColorMode() == utils.ColorModeLight, t, "Call to SetColorMode(utils.ColorModeLight) failed")

	user.SetColorMode(utils.ColorModeDark)

	testutils.Assert(user.GetColorMode() == utils.ColorModeDark, t, "Call to SetColorMode(utils.ColorModeDark) failed")

	pw := "password"
	user.SetPassword(pw)

	testutils.Assert(user.VerifyPassword(pw), t, "User password verification failed")

	width := 11
	height := 12
	user.SetWindowSize(width, height)

	testWidth, testHeight := user.WindowSize()

	testutils.Assert(testWidth == width && testHeight == height, t, "Call to SetWindowSize() failed")

	terminalType := "fake terminal type"
	user.SetTerminalType(terminalType)

	testutils.Assert(terminalType == user.TerminalType(), t, "Call to SetTerminalType() failed")
}

func Test_Character(t *testing.T) {
	character := database.NewCharacter("testcharacter", "", "")
	fakeId := bson.ObjectId("12345")

	testutils.Assert(character.IsNpc(), t, "Character with no userId should be an NPC")
	testutils.Assert(!character.IsPlayer(), t, "Character with no userId should NOT be a Player")
	testutils.Assert(character.IsOnline(), t, "NPCs should always be considered as online")

	character.SetUserId(fakeId)

	testutils.Assert(character.GetUserId() == fakeId, t, "Call to character.SetUser() failed", fakeId, character.GetUserId())
	testutils.Assert(!character.IsNpc(), t, "Character with a userId should NOT be an NPC")
	testutils.Assert(character.IsPlayer(), t, "Character with a userId should be a Player")
	testutils.Assert(!character.IsOnline(), t, "Player-Characters should be offline by default")

	character.SetOnline(true)

	testutils.Assert(character.IsOnline(), t, "Call to character.SetOnline(true) failed")

	character.SetRoomId(fakeId)

	testutils.Assert(character.GetRoomId() == fakeId, t, "Call to character.SetRoom() failed", fakeId, character.GetRoomId())

	cashAmount := 1234
	character.SetCash(cashAmount)

	testutils.Assert(character.GetCash() == cashAmount, t, "Call to character.GetCash() failed", cashAmount, character.GetCash())

	character.AddCash(cashAmount)

	testutils.Assert(character.GetCash() == cashAmount*2, t, "Call to character.AddCash() failed", cashAmount*2, character.GetCash())

	item1 := database.NewItem("test_item1")
	item2 := database.NewItem("test_item2")

	character.AddItem(item1)

	testutils.Assert(character.HasItem(item1), t, "Call to character.AddItem()/HasItem() failed - item1")

	character.AddItem(item2)

	testutils.Assert(character.HasItem(item2), t, "Call to character.AddItem()/HasItem() failed - item2")
	testutils.Assert(character.HasItem(item1), t, "Lost item1 after adding item2")

	conversation := "this is a fake conversation that is made up for the unit test"

	character.SetConversation(conversation)

	testutils.Assert(character.GetConversation() == conversation, t, "Call to character.SetConversation() failed")

	health := 123

	character.SetHealth(health)

	testutils.Assert(character.GetHealth() == health, t, "Call to character.SetHealth() failed")

	hitpoints := health - 10

	character.SetHitPoints(hitpoints)

	testutils.Assert(character.GetHitPoints() == hitpoints, t, "Call to character.SetHitPoints() failed")

	character.SetHitPoints(health + 10)

	testutils.Assert(character.GetHitPoints() == health, t, "Shouldn't be able to set a character's hitpoints to be greater than its maximum health", health, character.GetHitPoints())

	character.SetHealth(character.GetHealth() - 10)

	testutils.Assert(character.GetHitPoints() == character.GetHealth(), t, "Lowering health didn't lower the hitpoint count along with it", character.GetHitPoints(), character.GetHealth())

	character.SetHealth(100)
	character.SetHitPoints(100)

	hitAmount := 51
	character.Hit(hitAmount)

	testutils.Assert(character.GetHitPoints() == character.GetHealth()-hitAmount, t, "Call to character.Hit() failed", hitAmount, character.GetHitPoints())

	character.Heal(hitAmount)

	testutils.Assert(character.GetHitPoints() == character.GetHealth(), t, "Call to character.Heal() failed", hitAmount, character.GetHitPoints())
}

func Test_Zone(t *testing.T) {
	zoneName := "testzone"
	zone := database.NewZone(zoneName)

	testutils.Assert(zone.GetName() == zoneName, t, "Zone didn't have correct name upon creation", zoneName, zone.GetName())
}

func Test_Room(t *testing.T) {
	fakeZoneId := bson.ObjectId("!2345")
	room := database.NewRoom(fakeZoneId, database.Coordinate{X: 0, Y: 0, Z: 0})

	testutils.Assert(room.GetZoneId() == fakeZoneId, t, "Room didn't have correct zone ID upon creation", fakeZoneId, room.GetZoneId())

	fakeZoneId2 := bson.ObjectId("11111")
	room.SetZoneId(fakeZoneId2)
	testutils.Assert(room.GetZoneId() == fakeZoneId2, t, "Call to room.SetZoneId() failed")

	directionList := make([]database.ExitDirection, 10)
	directionCount := 10

	for i := 0; i < directionCount; i++ {
		directionList[i] = database.ExitDirection(i)
	}

	for _, dir := range directionList {
		testutils.Assert(!room.HasExit(dir), t, "Room shouldn't have any exits enabled by default", dir)
		room.SetExitEnabled(dir, true)
		testutils.Assert(room.HasExit(dir), t, "Call to room.SetExitEnabled(true) failed")
		room.SetExitEnabled(dir, false)
		testutils.Assert(!room.HasExit(dir), t, "Call to room.SetExitEnabled(false) failed")
	}

	item1 := database.NewItem("test_item1")
	item2 := database.NewItem("test_item2")

	room.AddItem(item1)
	testutils.Assert(room.HasItem(item1), t, "Call to room.AddItem(item1) faled")
	testutils.Assert(!room.HasItem(item2), t, "Room shouldn't have item2 in it yet")

	room.AddItem(item2)
	testutils.Assert(room.HasItem(item2), t, "Call to room.AddItem(item2) failed")
	testutils.Assert(room.HasItem(item1), t, "Room should still have item1 in it")

	room.RemoveItem(item1)
	testutils.Assert(!room.HasItem(item1), t, "Call to room.RemoveItem(item1) failed")
	testutils.Assert(room.HasItem(item2), t, "Room should still have item2 in it")

	room.RemoveItem(item2)
	testutils.Assert(!room.HasItem(item2), t, "Call to room.RemoveItem(item2) failed")
	testutils.Assert(!room.HasItem(item1), t, "Room still shouldn't have item1 in it")

	title := "Test Title"
	room.SetTitle(title)
	testutils.Assert(title == room.GetTitle(), t, "Call to room.SetTitle() failed", title, room.GetTitle())

	description := "This is a fake description"
	room.SetDescription(description)
	testutils.Assert(description == room.GetDescription(), t, "Call to room.SetDescription() failed", description, room.GetDescription())

	coord := database.Coordinate{X: 1, Y: 2, Z: 3}
	room.SetLocation(coord)
	testutils.Assert(coord == room.GetLocation(), t, "Call to room.SetLocation() failed", coord, room.GetLocation())
}

func Test_Item(t *testing.T) {
	name := "test_item"
	item := database.NewItem(name)

	testutils.Assert(item.GetName() == name, t, "Item didn't get created with correct name", name, item.GetName())
}

// vim: nocindent
