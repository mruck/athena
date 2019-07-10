package database

import (
	"fmt"
	"runtime"

	"gopkg.in/mgo.v2"
)

const MongoDbPort = "27017"

// getMongoHost returns the host platform for connecting to mongodb.
// Useful to tell if we are on k8s or local
func getMongoHost() string {
	if runtime.GOOS == "linux" {
		return "mongodb-service"
	}
	if runtime.GOOS == "darwin" {
		return "localhost"
	}
	panic("Unsupported OS")
}

func MustGetDatabase(port string, database string) *mgo.Database {
	host := getMongoHost()
	target := host + ":" + port
	//TODO: Add context timeout
	session, err := mgo.Dial(target)
	if err != nil {
		err = fmt.Errorf("unable to connect to mongodb server, is it running?\n"+
			"If testing start a mongo container with:\n"+
			"docker run -d --rm -p 27017:27017 mongo:3.4-xenial\n"+
			"%v", err)
		panic(err)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return session.DB(database)
}
