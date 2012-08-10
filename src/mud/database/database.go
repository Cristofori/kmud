package db

import (
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)

func FindUser( session *mgo.Session, name string ) bool {
    c := session.DB("mud").C("users")
    q := c.Find(bson.M{"name": name})

    count, err := q.Count()

    if err != nil {
        return false
    }

    return count > 0
}
