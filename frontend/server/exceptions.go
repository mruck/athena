package server

import "gopkg.in/mgo.v2"

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
	return nil, nil
}

func (manager *ExceptionsManager) ReadOne(filter interface{}, output interface{}) error {
	return manager.collection.Find(filter).One(output)
}

func (manager *ExceptionsManager) WriteOne(exc Exception) error {
	return manager.collection.Insert(exc)
}
