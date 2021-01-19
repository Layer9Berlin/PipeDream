package custom_strings

import (
	"fmt"
	"strings"
)

func QuoteValue(value string, quoteType string) string {
	switch quoteType {
	case "single":
		return fmt.Sprintf("'%v'", strings.Replace(value, "'", "\"'\"", -1))
	case "none":
		return value
	default:
		return fmt.Sprintf("\"%v\"", strings.Replace(value, "\"", "\\\"", -1))
	}
}
