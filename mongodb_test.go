package next

import (
    "testing"
    "log"
    "gopkg.in/mgo.v2/bson"
)

type People struct {
    Uid       string    `bson:"uid" json:"id"`
    Name      string    `bson:"name" json:"id"`
}


var mongo *MongoDB

func init() {
    mongo = NewMongoDB()
    mongo.Open("localhost", "test")
}

func TestInsertOrUpdate(t *testing.T) {
    p := People{"1", "Ace"}
    r := People{}
    err := mongo.InsertOrUpdate("test", "uid", "1", &p, &r)
    if err != nil {
        t.Error("Insert failed")
        log.Print(err)
        return
    }

    s := mongo.Session()
    c := s.DB("test").C("test")
    n, err := c.Find(bson.M{"uid": "1"}).Count()

    if err != nil {
        t.Error("Find failed")
        log.Print(err)
        return
    }

    if n != 1 {
        t.Error("Verify record failed")
        return
    }

    err = mongo.InsertOrUpdate("test", "uid", "1", &p, &r)

    if err != nil {
        t.Error("Find failed")
        log.Print(err)
        return
    }

    if n != 1 {
        t.Error("Verify record failed")
        return
    }

    t.Log("Mongodb InsertOrUpdate success")
}