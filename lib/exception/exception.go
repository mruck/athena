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
	FilePath   string
}

const Path = "/tmp/results/exceptions.json"

// NewExceptionsManager takes a connection to a mongo db and connects to the
// exceptions collection.  It also takes a path to a file to read contents from to
// write to the db.  If the path is the empty string, nothing shall be written to the db,
// and it will only be read from.
func NewExceptionsManager(db *mgo.Database, path string) *ExceptionsManager {
	return &ExceptionsManager{
		collection: db.C("exceptions"),
		FilePath:   path,
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
	exception, err := manager.ReadExceptionsFile()
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
func (manager *ExceptionsManager) ReadExceptionsFile() (*Exception, error) {
	// There's no file to read from
	if manager.FilePath == "" {
		return nil, nil
	}
	// Check if any exceptions were written
	empty, err := util.FileIsEmpty(manager.FilePath)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, nil
	}
	exception := &Exception{}
	err = util.UnmarshalFile(manager.FilePath, exception)
	return exception, err
}
