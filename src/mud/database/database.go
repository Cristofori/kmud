package database

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
    "mud/utils"
    //"fmt"
)

type dbError struct {
	message string
}

func (self dbError) Error() string {
	return self.message
}

func newDbError(message string) dbError {
	var err dbError
	err.message = message
	return err
}

type collectionName string

func getCollection(session *mgo.Session, collection collectionName) *mgo.Collection {
	return session.DB("mud").C(string(collection))
}

// Collection names
const (
    cUsers = collectionName("users")
    cCharacters = collectionName("characters")
)

// Field names
const (
    fName = "name"
    fCharacters = "characters"
)

func FindUser(session *mgo.Session, name string) (bool, error) {
	c := getCollection(session, cUsers)
	q := c.Find(bson.M{fName: name})

	count, err := q.Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func FindCharacter(session *mgo.Session, name string) (bool, error) {
	c := getCollection(session, cCharacters)
	q := c.Find(bson.M{fName: name})

	count, err := q.Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func NewUser(session *mgo.Session, name string) error {

	found, err := FindUser(session, name)

	if err != nil {
		return err
	}

	if found {
		return newDbError("That user already exists")
	}

	c := getCollection(session, cUsers)
	c.Insert(bson.M{fName: name})

	return nil
}

func NewCharacter(session *mgo.Session, user string, character string) error {

    found, err := FindCharacter(session, character)

    if err != nil {
        return err
    }

    if found {
        return newDbError("That character already exists")
    }

	c := getCollection(session, cUsers)
	c.Update(bson.M{fName: user}, bson.M{"$push": bson.M{fCharacters: character}})

    c = getCollection(session, cCharacters)
    c.Insert(bson.M{fName: character})

	return nil
}

func GetCharacterLocation(session *mgo.Session, name string) (string, error) {
	c := getCollection(session, cCharacters)
	q := c.Find(bson.M{fName: name})

	result := map[string]string{}
	err := q.One(&result)

	if err != nil {
		return "", err
	}

	return result["location"], nil
}

func SetUserLocation(session *mgo.Session, name string, locationId string) error {
	c := getCollection(session, cUsers)
	c.Update(bson.M{fName: name}, bson.M{"location": locationId})
	return nil
}

func GetUserCharacters(session *mgo.Session, name string) ([]string, error) {
	c := getCollection(session, cUsers)
	q := c.Find(bson.M{fName: name})

	result := map[string][]string{}
	err := q.One(&result)

	return result[fCharacters], err
}

func DeleteCharacter(session *mgo.Session, user string, character string) error {
	c := getCollection(session, cUsers)
	c.Update(bson.M{fName: user}, bson.M{"$pull": bson.M{fCharacters: utils.Simplify(character)}})

    c = getCollection(session, cCharacters)
    c.Remove(bson.M{fName: character})

	return nil
}

// vim: nocindent
