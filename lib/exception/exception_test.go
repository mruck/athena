package exception

// To run this test you need mongo up and running:
// docker run -d  -p 27017:27017 mongo

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/moul/http2curl"
	"github.com/mruck/athena/lib/database"
	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

func TestWriteReadOne(t *testing.T) {
	db := database.MustGetDatabase(database.MongoDbPort, "test")
	exceptions := NewExceptionsManager(db, "")
	_ = exceptions.Drop()
	exn := Exception{"get", "/test/route", "InvalidRead", "Test Mesage", "12345", "fake curl cmd"}
	err := exceptions.WriteOne(exn)
	require.NoError(t, err)
	result, err := exceptions.ReadOne("12345")
	require.NoError(t, err)
	require.Equal(t, "get", result.Method)
	require.Equal(t, "/test/route", result.Path)
	require.Equal(t, "InvalidRead", result.Class)
	require.Equal(t, "fake curl cmd", result.Curl)
}

func TestWriteReadAll(t *testing.T) {
	db := database.MustGetDatabase(database.MongoDbPort, "test")
	exceptions := NewExceptionsManager(db, "")
	_ = exceptions.Drop()
	exn := Exception{"get", "/test/route", "InvalidRead", "Test Mesage", "12345", "fake curl cmd"}
	err := exceptions.WriteOne(exn)
	require.NoError(t, err)
	result, err := exceptions.GetAll("12345")
	require.NoError(t, err)
	require.Equal(t, "get", result[0].Method)
	require.Equal(t, "/test/route", result[0].Path)
	require.Equal(t, "InvalidRead", result[0].Class)

	exn = Exception{"get2", "/test/route2", "InvalidRead2", "Test Mesage2", "12345", "fake curl cmd"}
	err = exceptions.WriteOne(exn)
	require.NoError(t, err)
	results, err := exceptions.GetAll("12345")
	require.NoError(t, err)
	for _, result := range results {
		log.Info(result)
	}
}

func TestUpdate(t *testing.T) {
	// Create a dummy exceptions file
	method := "get"
	path := "/info"
	targetid1 := "targetid1"
	class := "NoMethodError"
	message := "There's no method for this"

	// Create a dummy request
	req, err := http.NewRequest("GET", "/info", nil)
	require.NoError(t, err)
	curl, err := http2curl.GetCurlCommand(req)
	require.NoError(t, err)

	// Dummy exception
	exn1 := Exception{
		Method:   method,
		Path:     path,
		Class:    class,
		Message:  message,
		TargetID: targetid1,
	}

	// Marshal to file
	tmp, err := ioutil.TempFile("/tmp", "")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	err = util.MarshalToFile(exn1, tmp.Name())
	require.NoError(t, err)

	// Connect to mongodb to log exceptions
	db := database.MustGetDatabase(database.MongoDbPort, "testdb")
	manager := NewExceptionsManager(db, tmp.Name())

	// Drop the table from previous tests
	_ = manager.Drop()

	// Update exceptions table by reading from the mock file
	err = manager.Update(path, method, targetid1, curl)
	require.NoError(t, err)

	// Update dummy exceptions file by writing the same
	// stuff with a new target id
	targetid2 := "targetid2"
	exn2 := Exception{
		Method:   method,
		Path:     path,
		Class:    class,
		Message:  message,
		TargetID: targetid2,
	}
	err = util.MarshalToFile(exn2, tmp.Name())
	require.NoError(t, err)

	// Update again
	err = manager.Update(path, method, targetid2, curl)
	require.NoError(t, err)

	// Truncate the exceptions file so its empty
	err = os.Truncate(tmp.Name(), 0)
	require.NoError(t, err)

	// Update again
	err = manager.Update(path, method, targetid2, curl)
	require.NoError(t, err)

	// Check our results for targetid1
	result, err := manager.ReadOne(targetid1)
	require.NoError(t, err)
	require.Equal(t, exn1.Method, result.Method)
	require.Equal(t, exn1.Path, result.Path)
	require.Equal(t, exn1.Class, result.Class)
	require.Equal(t, exn1.TargetID, result.TargetID)
	require.Equal(t, exn1.Message, result.Message)
	require.Equal(t, curl.String(), result.Curl)

	// Check our results targetid2
	result, err = manager.ReadOne(targetid2)
	require.NoError(t, err)
	require.Equal(t, exn2.Method, result.Method)
	require.Equal(t, exn2.Path, result.Path)
	require.Equal(t, exn2.Class, result.Class)
	require.Equal(t, exn2.TargetID, result.TargetID)
	require.Equal(t, curl.String(), result.Curl)
}
