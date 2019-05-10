package swagger

import (
	"testing"

	"github.com/mruck/athena/goFuzz/util"
)

func TestReadSwagger(t *testing.T) {
	swagger := ReadSwagger("dummySwagger.json")
	util.PrettyPrintStruct(swagger)
}
