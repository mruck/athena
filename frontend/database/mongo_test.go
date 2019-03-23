package database

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"
)

type Exception struct {
	Verb     string `bson:"Verb"`
	Path     string `bson:"Path"`
	Class    string `bson:"Class"`
	Message  string `bson:"Message"`
	TargetID string `bson:"TargetID"`
}

func TestWriteRead(t *testing.T) {
	cli, err := NewClient("localhost", "27017", "test")
	require.NoError(t, err)
	require.NotNil(t, cli)

	tablename := uuid.New().String()
	fmt.Println(tablename)

	err = cli.WriteOne(tablename, &Exception{"put", "/test", "InvalidWrite", "This is a test message", "123456"})
	require.NoError(t, err)

	var result Exception
	err = cli.ReadOne(tablename, bson.M{"TargetID": "123456"}, &result)
	require.NoError(t, err)
	require.Equal(t, "put", result.Verb)
	require.Equal(t, "/test", result.Path)
	require.Equal(t, "InvalidWrite", result.Class)
	require.Equal(t, "This is a test message", result.Message)
}

func TestMongoClient(t *testing.T) {
	cli, err := NewClient("localhost", "27017", "athena")
	require.NoError(t, err)
	require.NotNil(t, cli)
}
