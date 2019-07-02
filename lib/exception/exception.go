package exception

import (
	"io/ioutil"
	"os"

	"github.com/mruck/athena/lib/database"
	"github.com/mruck/athena/lib/log"
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

// ExceptionsManager tracks exceptions in memory and logs them to a db
type ExceptionsManager struct {
	collection *mgo.Collection
	filePath   string
	// Keep track of exceptions in memory as well
	uniqueExceptions []Exception
	// Did we see a new exception?
	Delta bool
}

const Path = "/tmp/results/exceptions.json"

// NewExceptionsManager takes a connection to a mongo db and connects to the
// exceptions collection.  It also takes a path to a file to read contents from to
// write to the db.  If the path is the empty string, nothing shall be written to the db,
// and it will only be read from.
func NewExceptionsManager(db *mgo.Database, path string) *ExceptionsManager {
	manager := &ExceptionsManager{
		collection: db.C("exceptions"),
		filePath:   path,
	}

	// We may have run on this target before.  If so, reload the exceptions
	// that we've seen before
	targetID := util.DefaultEnv("TARGET_ID", "")
	if targetID != "" {
		exceptions, err := manager.GetAll(targetID)
		if err != nil {
			log.Fatal(err)
		}
		manager.uniqueExceptions = exceptions
	}
	return manager

}

func (manager *ExceptionsManager) GetAll(targetID string) ([]Exception, error) {
	var results []Exception
	query := bson.M{"TargetID": targetID}
	iter := manager.collection.Find(query).Limit(100).Iter()
	err := iter.All(&results)
	return results, errors.WithStack(err)
}

// ReadOne reads a single exception by target id
func (manager *ExceptionsManager) ReadOne(targetID string) (Exception, error) {
	var result Exception
	query := bson.M{"TargetID": targetID}
	err := manager.collection.Find(query).One(&result)
	return result, err
}

// WriteOne writes a single exception
func (manager *ExceptionsManager) WriteOne(exc Exception) error {
	return errors.WithStack(manager.collection.Insert(exc))
}

// Drop a collection
func (manager *ExceptionsManager) Drop() error {
	return errors.WithStack(manager.collection.DropCollection())
}

// return whether or not an exception is benign
func (exception *Exception) benign() bool {
	return false
}

func exceptionsEqual(exn1 Exception, exn2 Exception) bool {
	return exn1.Path == exn2.Path &&
		exn1.Method == exn2.Method &&
		exn1.Class == exn2.Class
}

// Update exceptions database from exceptions written by rails
func (manager *ExceptionsManager) Update(path string, method string, targetid string) error {
	// Assume we don't see a unique exception
	manager.Delta = false

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

	// Add extra metadata to the exception
	exception.Path = path
	exception.Method = method
	exception.TargetID = targetid

	// Have we seen this exception before?
	for _, oldException := range manager.uniqueExceptions {
		// We've already logged this exception
		if exceptionsEqual(oldException, *exception) {
			return nil
		}

	}

	// This exception is unique
	manager.Delta = true
	manager.uniqueExceptions = append(manager.uniqueExceptions, *exception)

	// Log to db
	return manager.WriteOne(*exception)
}

// ReadExceptionsFile reads the file written by rails logging exceptions
func (manager *ExceptionsManager) ReadExceptionsFile() (*Exception, error) {
	// There's no file to read from
	if manager.filePath == "" {
		return nil, nil
	}
	// Check if any exceptions were written
	empty, err := util.FileIsEmpty(manager.filePath)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, nil
	}
	exception := &Exception{}
	err = util.UnmarshalFile(manager.filePath, exception)
	return exception, err
}

// ReadDB connects to athena db and reads the TARGET_ID exceptions table.
// To be called inside a pod where TARGET_ID is set and there's a running
// mongo db instance
func ReadDB() {
	// Get target id
	targetID := os.Getenv("TARGET_ID")
	if targetID == "" {
		log.Info("TARGET_ID not set.  Are you running inside a pod with athena set up?")
	}

	// Connect to db
	db := database.MustGetDatabase(database.MongoDbPort, "athena")
	manager := NewExceptionsManager(db, Path)

	// Read all exceptions
	exceptions, err := manager.GetAll(targetID)
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}

	// Log to tmp file
	tmpfile, err := ioutil.TempFile("", "exceptions_")
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}
	err = util.MarshalToFile(exceptions, tmpfile.Name())
	if err != nil {
		log.Fatal(errors.WithStack(err))
	}
	log.Infof("Logged exceptions to %s", tmpfile.Name())
}
