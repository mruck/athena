package har

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalHar(t *testing.T) {
	har := UnmarshalHar("../tests/login_har.json")
	// Pick something random to check for equality
	request0 := har.Log.Entries[0].Request
	require.Equal(t, "http://localhost:50121/session/csrf?_=1548444062137", request0.URL)
	//util.PrettyPrint(har)
}
