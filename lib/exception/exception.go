package exception

import (
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Exception datatype
type Exception struct {
	Method   string `bson:"Verb"`
	Path     string `bson:"Path"`
	Class    string `bson:"Class"`
	Message  string `bson:"Message"`
	TargetID string `bson:"TargetID"`
}

type ExceptionsManager struct {
	collection *mgo.Collection
}

const exceptionsFile = "/tmp/results/exceptions.json"

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
	return errors.WithStack(manager.collection.Insert(exc))
}

func (manager *ExceptionsManager) Drop() error {
	return errors.WithStack(manager.collection.DropCollection())
}

// return whether or not an exception is benign
func (exception *Exception) benign() bool {
	return false
}

// Update exceptions database from exceptions written by rails
func (manager *ExceptionsManager) Update(path string, method string, targetid string) error {
	exception, err := ReadExceptionsFile()
	if err != nil {
		return err
	}
	// There was no exception
	if exception == nil {
		return nil
	}
	if exception.benign() {
		return nil
	}
	exception.Path = path
	exception.Method = method
	exception.TargetID = targetid
	return manager.WriteOne(*exception)
}

// ReadExceptionsFile reads the file written by rails logging exceptions
func ReadExceptionsFile() (*Exception, error) {
	// Check if any exceptions were written
	empty, err := util.FileIsEmpty(exceptionsFile)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, nil
	}
	exception := &Exception{}
	err = util.UnmarshalFile(exceptionsFile, exception)
	return exception, err
}
