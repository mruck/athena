package database

import (
	"gopkg.in/mgo.v2"
)

// Client to MongoDB
type Client struct {
	database *mgo.Database
}

// NewClient creates a connection to the target db
func NewClient(host string, port string, database string) (*Client, error) {
	// Get a client
	target := host + ":" + port
	//TODO: Add context timeout
	session, err := mgo.Dial(target)
	if err != nil {
		return nil, err
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return &Client{session.DB(database)}, nil
}

// WriteOne writes one entry to given table.
func (c *Client) WriteOne(collectionName string, document interface{}) error {
	collection := c.database.C(collectionName)
	return collection.Insert(document)
}

// ReadOne reads one entry from given table.
func (c *Client) ReadOne(collectionName string, filter interface{}, output interface{}) error {
	collection := c.database.C(collectionName)
	return collection.Find(filter).One(output)
}
