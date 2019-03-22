package database

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

type exception struct {
	Verb     string
	Path     string
	Class    string
	Message  string
	TargetID string
}

func (ex exception) ToBSON() bson.M {
	return bson.M{
		"Verb": ex.Verb,
		"Path": ex.Path,
	}
}

func TestSanity(t *testing.T) {

}

func TestMongoClient(t *testing.T) {
	//	t.Run("connect", func(t *testing.T) {
	//		cli, err := NewClient("localhost", "27017", "athena")
	//		require.NoError(t, err)
	//		require.NotNil(t, cli)
	//	})

	t.Run("put then get", func(t *testing.T) {
		cli, err := NewClient("localhost", "27017", "athena")
		require.NoError(t, err)
		require.NotNil(t, cli)

		//		exc := exception{Verb: "v", Path: "p"}
		//		err = cli.WriteOne("exceptions", exc)
		//		require.NoError(t, err)
		//
		//		result := exception{}
		//		err = cli.ReadOne("exceptions", bson.M{"Verb": "v", "Path": "p"}, &result)
		//		require.NoError(t, err)
		//		require.Equal(t, "v", result.Verb)
		//		require.Equal(t, "p", result.Path)
	})
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
}
