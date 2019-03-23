package database

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

type Exception struct {
	Verb     string `bson:"Verb"`
	Path     string `bson:"Path"`
	Class    string `bson:"Class"`
	Message  string `bson:"Message"`
	TargetID string `bson:"TargetID"`
}

func TestWriteRead(t *testing.T) {
	cli, err := NewClient("localhost", "27017", "athena")
	require.NoError(t, err)
	require.NotNil(t, cli)

	err = cli.WriteOne("exceptions", &Exception{"put", "/test", "InvalidWrite", "This is a test message", "123456"})
	require.NoError(t, err)

	var result Exception
	err = cli.ReadOne("exceptions", bson.M{"TargetID": "123456"}, &result)
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

//	t.Run("put then get", func(t *testing.T) {
//		cli, err := NewClient("localhost", "27017", "athena")
//		require.NoError(t, err)
//		require.NotNil(t, cli)

//		exc := exception{Verb: "v", Path: "p"}
//		err = cli.WriteOne("exceptions", exc)
//		require.NoError(t, err)
//
//		result := exception{}
//		err = cli.ReadOne("exceptions", bson.M{"Verb": "v", "Path": "p"}, &result)
//		require.NoError(t, err)
//		require.Equal(t, "v", result.Verb)
//		require.Equal(t, "p", result.Path)
//	})
//	// read exception inserted from python
//	t.Run("read exception", func(t *testing.T) {
//		cli, err := NewClient("localhost", "27017", "athena")
//		require.NoError(t, err)
//		require.NotNil(t, cli)
//		result := exception{}
//		err = cli.ReadOne("exceptions", bson.M{"target_id": "2bedb7f5501f4b4393b479e3a5e91bf6"}, &result)
//		require.NoError(t, err)
//		require.Equal(t, result.Verb, "put")
//		require.Equal(t, result.Path, "/this/is/a/test/route")
//	})
