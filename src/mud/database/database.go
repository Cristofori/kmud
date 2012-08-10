package database

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type dbError struct {
    message string
}

func (self dbError) Error() string {
    return self.message
}

func newDbError( message string ) dbError {
    var err dbError
    err.message = message
    return err
}

func FindUser(session *mgo.Session, name string) bool {
	c := session.DB("mud").C("users")
	q := c.Find(bson.M{"name": name})

	count, err := q.Count()

	if err != nil {
		return false
	}

	return count > 0
}

func NewUser(session *mgo.Session, name string) error {
	if FindUser(session, name) {
		return newDbError("That user already exists")
	}

	c := session.DB("mud").C("users")
	c.Insert(bson.M{"name": name})

	return nil
}

// vim: nocindent
