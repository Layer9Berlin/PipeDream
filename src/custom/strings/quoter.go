package strings

import (
	"fmt"
	systemstrings "strings"
)

// QuoteValue adds surrounding quotes around a string
//
// Available quoteTypes are 'single', 'none', and 'double' (default)
func QuoteValue(value string, quoteType string) string {
	switch quoteType {
	case "single":
		return fmt.Sprintf("'%v'", systemstrings.Replace(value, "'", "\"'\"", -1))
	case "none":
		return value
	default:
		return fmt.Sprintf("\"%v\"", systemstrings.Replace(value, "\"", "\\\"", -1))
	}
}
