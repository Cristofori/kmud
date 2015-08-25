package database

import (
	"gopkg.in/mgo.v2"
)

func NewMongoSession(session *mgo.Session) *MongoSession {
	var mongoSession MongoSession
	mongoSession.session = session
	return &mongoSession
}

type MongoSession struct {
	session *mgo.Session
}

func (ms MongoSession) DB(dbName string) Database {
	var db MongoDatabase
	db.database = ms.session.DB(dbName)
	return &db
}

type MongoDatabase struct {
	database *mgo.Database
}

func (md MongoDatabase) C(collectionName string) Collection {
	return &MongoCollection{collection: md.database.C(collectionName)}
}

type MongoCollection struct {
	collection *mgo.Collection
}

func (mc MongoCollection) FindId(id interface{}) Query {
	return &MongoQuery{query: mc.collection.FindId(id)}
}

func (mc MongoCollection) Find(selector interface{}) Query {
	return &MongoQuery{query: mc.collection.Find(selector)}
}

func (mc MongoCollection) RemoveId(id interface{}) error {
	return mc.collection.RemoveId(id)
}

func (mc MongoCollection) Remove(selector interface{}) error {
	return mc.collection.Remove(selector)
}

func (mc MongoCollection) DropCollection() error {
	return mc.collection.DropCollection()
}

func (mc MongoCollection) UpdateId(id interface{}, change interface{}) error {
	return mc.collection.UpdateId(id, change)
}

func (mc MongoCollection) UpsertId(id interface{}, change interface{}) error {
	_, err := mc.collection.UpsertId(id, change)
	return err
}

type MongoQuery struct {
	query *mgo.Query
}

func (mq MongoQuery) Count() (int, error) {
	return mq.query.Count()
}

func (mq MongoQuery) One(result interface{}) error {
	return mq.query.One(result)
}

func (mq MongoQuery) Iter() Iterator {
	return &MongoIterator{iterator: mq.query.Iter()}
}

type MongoIterator struct {
	iterator *mgo.Iter
}

func (mi MongoIterator) All(result interface{}) error {
	return mi.iterator.All(result)
}

// vim: nocindent
