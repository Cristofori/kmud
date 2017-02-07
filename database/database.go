package database

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2/bson"
)

type Session interface {
	DB(string) Database
}

type Database interface {
	C(string) Collection
}

type Collection interface {
	Find(interface{}) Query
	FindId(interface{}) Query
	RemoveId(interface{}) error
	Remove(interface{}) error
	DropCollection() error
	UpdateId(interface{}, interface{}) error
	UpsertId(interface{}, interface{}) error
}

type Query interface {
	Count() (int, error)
	One(interface{}) error
	Iter() Iterator
}

type Iterator interface {
	All(interface{}) error
}

var modifiedObjects = map[types.Id]bool{}
var modifiedObjectChannel chan types.Id

var _session Session
var _dbName string

func init() {
	modifiedObjectChannel = make(chan types.Id, 1)
	watchModifiedObjects()
}

func Init(session Session, dbName string) {
	_session = session
	_dbName = dbName
}

func dbinit(obj types.Object) {
	obj.SetId(bson.NewObjectId())
	datastore.Set(obj)
	commitObject(obj.GetId())
}

func watchModifiedObjects() {
	go func() {
		timeout := make(chan bool)

		startTimeout := func() {
			go func() {
				time.Sleep(1 * time.Second)
				timeout <- true
			}()
		}

		startTimeout()

		for {
			select {
			case id := <-modifiedObjectChannel:
				modifiedObjects[id] = true
			case <-timeout:
				for id := range modifiedObjects {
					go commitObject(id)
				}
				modifiedObjects = map[types.Id]bool{}
				startTimeout()
			}
		}
	}()
}

func getCollection(collection types.ObjectType) Collection {
	return _session.DB(_dbName).C(string(collection))
}

func getCollectionOfObject(obj types.Object) Collection {
	name := reflect.TypeOf(obj).String()
	parts := strings.Split(name, ".")
	name = parts[len(parts)-1]

	return getCollection(types.ObjectType(name))
}

func Retrieve(id types.Id, typ types.ObjectType) types.Object {
	if datastore.ContainsId(id) {
		return datastore.Get(id)
	}

	var object types.Object

	switch typ {
	case types.PcType:
		object = &Pc{}
	case types.NpcType:
		object = &Npc{}
	case types.SpawnerType:
		object = &Spawner{}
	case types.UserType:
		object = &User{}
	case types.ZoneType:
		object = &Zone{}
	case types.AreaType:
		object = &Area{}
	case types.RoomType:
		object = &Room{}
	case types.TemplateType:
		object = &Template{}
	case types.ItemType:
		object = &Item{}
	case types.SkillType:
		object = &Skill{}
	case types.EffectType:
		object = &Effect{}
	case types.StoreType:
		object = &Store{}
	case types.WorldType:
		object = &World{}
	default:
		panic(fmt.Sprintf("unrecognized object type: %v", typ))
	}

	c := getCollectionOfObject(object)
	err := c.FindId(id).One(object)

	if err != nil || object == nil {
		return nil
	}

	datastore.Set(object)
	return object
}

func RetrieveObjects(t types.ObjectType, objects interface{}) {
	c := getCollection(t)
	err := c.Find(nil).Iter().All(objects)
	utils.HandleError(err)
}

func Find(t types.ObjectType, query bson.M) []bson.ObjectId {
	return find(t, query)
}

func FindOne(t types.ObjectType, query bson.M) types.Id {
	var result bson.M
	find_helper(t, query).One(&result)
	id, found := result["_id"]
	if found {
		return id.(bson.ObjectId)
	}
	return nil
}

func FindAll(t types.ObjectType) []bson.ObjectId {
	return find(t, nil)
}

func find(t types.ObjectType, query interface{}) []bson.ObjectId {
	var results []bson.M
	find_helper(t, query).Iter().All(&results)

	var ids []bson.ObjectId
	for _, result := range results {
		ids = append(ids, result["_id"].(bson.ObjectId))
	}

	return ids
}

func find_helper(t types.ObjectType, query interface{}) Query {
	c := getCollection(t)
	return c.Find(query)
}

func DeleteObject(id types.Id) {
	object := datastore.Get(id)
	datastore.Remove(object)

	object.Destroy()

	c := getCollectionOfObject(object)
	utils.HandleError(c.RemoveId(object.GetId()))
}

func commitObject(id types.Id) {
	object := datastore.Get(id)

	if object == nil || object.IsDestroyed() {
		return
	}

	c := getCollectionOfObject(object)

	object.ReadLock()
	err := c.UpsertId(object.GetId(), object)
	object.ReadUnlock()

	if err != nil {
		fmt.Println("Update failed", object.GetId())
	}

	utils.HandleError(err)
}
