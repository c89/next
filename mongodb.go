package next

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

type MongoDB struct {
    session *mgo.Session
    db string
}

func NewMongoDB() *MongoDB {
    return &MongoDB{}
}

func (mongo *MongoDB) Open(url string, db string) {
    s, err := mgo.Dial(url)
    if err != nil {
        panic(err.Error())
    }

    mongo.session = s
    mongo.db = db
}

func (mongo *MongoDB) Close() {
    mongo.session.Close()
}

func (mongo *MongoDB) Session() *mgo.Session {
    s := mongo.session.Copy()
    s.SetMode(mgo.Strong, true)
    return s
}

func (mongo *MongoDB) InsertOrUpdate(collection string, buziKey string, buziValue interface{}, d interface{}, result interface{}) (error) {
    s := mongo.Session()
    defer s.Close()

    c := s.DB(mongo.db).C(collection)

    change := mgo.Change{
        Update:     bson.M{"$set": d},
        Upsert:     true,
        Remove:     false,
        ReturnNew:  true,
    }

    _, err := c.Find(bson.M{buziKey: buziValue}).Apply(change, result)
    return err
}

// func (mongo *Mongodb) InsertOrUpdate(collection string, condition map[string]interface{}, document interface{}) (int, error) {
//     s := mongo.Session()
//     defer s.Close()

//     c := s.DB(mongo.db).C(collection)
//     n, err := c.FindId(id).Count()

//     if err != nil {
//         return 0, err
//     }

//     if n > 0 {
//         err = c.UpdateId(id, document)
//         if err != nil {
//             return 0, err
//         }
//     }  else {
//         err = c.Insert(document)
//         if err != nil {
//             return 0, err
//         }
//     }

//     return 1, nil
// }

// func (mongo *Mongodb) Get(collection string, id string, document interface{}) (error) {
//     s := mongo.Session()
//     defer s.Close()

//     c := s.DB(mongo.db).C(collection)
//     err := c.FindId(id).One(&document)
//     return err
// }

// func (mongo *Mongodb) Remove(collection string, id string) (error) {
//     s := mongo.Session()
//     defer s.Close()

//     c := s.DB(mongo.db).C(collection)
//     err := c.RemoveId(id)
//     return err
// }
