package database

import (
	"gopkg.in/mgo.v2"
)

func MustGetDatabase(host string, port string, database string) *mgo.Database {
	target := host + ":" + port
	//TODO: Add context timeout
	session, err := mgo.Dial(target)
	if err != nil {
		panic("Unable to connect to mongodb server, is it running? %v", err)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return session.DB(database)
}
