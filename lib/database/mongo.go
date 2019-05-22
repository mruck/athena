package database

import (
	"fmt"
	"runtime"

	"gopkg.in/mgo.v2"
)

const MongoDbPort = "27017"

//MustGetHost returns the host platform for connecting to mongodb. Useful to tell if we are on k8s or local
func MustGetHost() string {
	if runtime.GOOS == "linux" {
		return "mongodb-service"
	}
	if runtime.GOOS == "darwin" {
		return "localhost"
	}
	panic("Unsupported OS")
}

func MustGetDatabase(port string, database string) *mgo.Database {
	host := MustGetHost()
	target := host + ":" + port
	//TODO: Add context timeout
	session, err := mgo.Dial(target)
	if err != nil {
		err = fmt.Errorf("unable to connect to mongodb server, is it running? %v", err)
		panic(err)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return session.DB(database)
}
