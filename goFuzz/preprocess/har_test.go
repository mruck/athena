package preprocess

import (
	"testing"

	"github.com/mruck/athena/goFuzz/util"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalHar(t *testing.T) {
	har, err := unmarshalHar("test/login_har.json")
	require.NoError(t, err)
	util.PrettyPrint(har)
}
