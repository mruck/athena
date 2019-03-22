package database

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Client struct {
	database *mgo.Database
}

func NewClient(host string, port string, database string) (*Client, error) {
	// Get a client
	target := host + ":" + port
	session, err := mgo.Dial(target)
	if err != nil {
		return nil, err
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return &Client{session.DB(database)}, nil
}

type Bsonable interface {
	ToBSON() bson.M
}

// WriteOne writes one entry to given table.
func (c *Client) WriteOne(table string, document Bsonable) error {
	return nil
}

// ReadOne reads one entry from given table.
func (c *Client) ReadOne(table string, filter bson.M, output interface{}) error {
	return nil
}
