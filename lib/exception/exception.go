package exception

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Exception datatype
type Exception struct {
	Verb     string `bson:"Verb"`
	Path     string `bson:"Path"`
	Class    string `bson:"Class"`
	Message  string `bson:"Message"`
	TargetID string `bson:"TargetID"`
}

type ExceptionsManager struct {
	collection *mgo.Collection
}

func NewExceptionsManager(db *mgo.Database) *ExceptionsManager {
	return &ExceptionsManager{
		collection: db.C("exceptions"),
	}
}

func (manager *ExceptionsManager) GetAll(targetID string) ([]Exception, error) {
	var results []Exception
	query := bson.M{"TargetID": targetID}
	iter := manager.collection.Find(query).Limit(100).Iter()
	err := iter.All(&results)
	return results, err
}

// ReadOne reads a single exception by target id
func (manager *ExceptionsManager) ReadOne(targetID string) (Exception, error) {
	var result Exception
	query := bson.M{"TargetID": targetID}
	err := manager.collection.Find(query).One(&result)
	return result, err
}

func (manager *ExceptionsManager) WriteOne(exc Exception) error {
	return manager.collection.Insert(exc)
}

func (manager *ExceptionsManager) Drop() error {
	return manager.collection.DropCollection()
}
