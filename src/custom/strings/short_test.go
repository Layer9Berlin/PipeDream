package strings

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStrings_Shorten(t *testing.T) {
	require.Equal(t, "extrem…", Shorten("extremely long string 1234732 237482 347657943543534657 3240524523", 6))
	require.Equal(t, "test1⇤\xef\xb8…", Shorten("test1\r\ntest2\r\n", 10))
}
