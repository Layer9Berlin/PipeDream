package strings

import (
	"fmt"
	nativestrings "strings"
)

// Shorten turns a string into a single-line representation of at most maxLength bytes
func Shorten(commandString string, maxLength int) string {
	commandString = nativestrings.Replace(commandString, "\n", "↩", -1)
	commandString = nativestrings.Replace(commandString, "\r", "⇤︎", -1)
	if len(commandString) > maxLength {
		return fmt.Sprintf("%v…", commandString[:maxLength])
	}
	return commandString
}
