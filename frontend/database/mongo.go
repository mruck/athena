package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Client struct {
	database *mongo.Database
}

func NewClient(host string, port string, database string) (*Client, error) {
	// Get a client
	target := "mongodb://" + host + ":" + port
	client, err := mongo.NewClient(options.Client().ApplyURI(target))
	if err != nil {
		return nil, err
	}
	// Connect
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	// Ping
	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	// Connect to the db
	return &Client{client.Database(database)}, nil
}

type Bsonable interface {
	ToBSON() bson.M
}

// WriteOne writes one entry to given table.
func (c *Client) WriteOne(table string, document Bsonable) error {
	collection := c.database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	value := document.ToBSON()
	_, err := collection.InsertOne(ctx, value)
	if err != nil {
		fmt.Println("Error inserting")
		return err
	}
	return nil
}

// ReadOne reads one entry from given table.
func (c *Client) ReadOne(table string, filter bson.M, output interface{}) error {
	collection := c.database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	result := collection.FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		return err
	}
	return result.Decode(output)
}
