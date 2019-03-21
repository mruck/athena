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

type exception struct {
	verb      string
	path      string
	class     string
	message   string
	target_id string
}

func (c *Client) LookUp(table string) (string, error) {
	collection := c.database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	filter := bson.M{"target_id": "5839485c36b54f87a9ece210ec4943e8"}
	var result exception
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return "", err
	}
	fmt.Println(result)
	return "", nil
}
