package server

// To run this test you need mongo up and running:
// docker run -p 27017:27017 mongo

import (
	"fmt"
	"testing"

	"github.com/mruck/athena/frontend/database"
	"github.com/stretchr/testify/require"
)

func TestWriteReadOne(t *testing.T) {
	db := database.MustGetDatabase("localhost", "27017", "test")
	exceptions := NewExceptionsManager(db)
	_ = exceptions.Drop()
	//require.NoError(t, err)
	exn := Exception{"get", "/test/route", "InvalidRead", "Test Mesage", "12345"}
	err := exceptions.WriteOne(exn)
	require.NoError(t, err)
	result, err := exceptions.ReadOne("12345")
	require.NoError(t, err)
	require.Equal(t, "get", result.Verb)
	require.Equal(t, "/test/route", result.Path)
	require.Equal(t, "InvalidRead", result.Class)
}

func TestWriteReadAll(t *testing.T) {
	db := database.MustGetDatabase("localhost", "27017", "test")
	exceptions := NewExceptionsManager(db)
	_ = exceptions.Drop()
	//require.NoError(t, err)
	exn := Exception{"get", "/test/route", "InvalidRead", "Test Mesage", "12345"}
	err := exceptions.WriteOne(exn)
	require.NoError(t, err)
	result, err := exceptions.GetAll("12345")
	require.NoError(t, err)
	require.Equal(t, "get", result[0].Verb)
	require.Equal(t, "/test/route", result[0].Path)
	require.Equal(t, "InvalidRead", result[0].Class)

	exn = Exception{"get2", "/test/route2", "InvalidRead2", "Test Mesage2", "12345"}
	err = exceptions.WriteOne(exn)
	require.NoError(t, err)
	results, err := exceptions.GetAll("12345")
	require.NoError(t, err)
	for _, result := range results {
		fmt.Println(result)
	}
}
