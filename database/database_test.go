package database

import (
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

func (ms TestSession) DB(dbName string) Database {
	return &TestDatabase{}
}

type TestDatabase struct {
}

func (md TestDatabase) C(collectionName string) Collection {
	return &TestCollection{}
}

type TestCollection struct {
}

func (mc TestCollection) Find(selector interface{}) Query {
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

func (mq TestQuery) Iter() Iterator {
	return &TestIterator{}
}

type TestIterator struct {
}

func (mi TestIterator) All(result interface{}) error {
	return nil
}

func Test_ThreadSafety(t *testing.T) {
	runtime.GOMAXPROCS(2)
	Init(&TestSession{})

	char := NewCharacter("test", "", "")

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
	user := NewUser("testuser", "")

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
	character := NewCharacter("testcharacter", "", "")
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

	item1 := NewItem("")
	item2 := NewItem("")

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
	// zone := NewZone("")
}

func Test_Room(t *testing.T) {
	// room := NewRoom("")
}

func Test_Item(t *testing.T) {
	// item := NewItem("")
}

// vim: nocindent
