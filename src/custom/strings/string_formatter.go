package strings

import (
	"fmt"
	"math"
	"strings"
)

func PrettyPrintedByteCount(byteCount int) string {
	if byteCount < 1000 {
		return fmt.Sprintf("%dB", byteCount)
	}
	exponent := math.Floor(math.Log(float64(byteCount)) / math.Log(1000))
	divisor := math.Pow(1000, exponent)
	return fmt.Sprintf("%.1f%cB", float64(byteCount)/divisor, "kMGTPE"[int(exponent-1)])
}

func IdentifierToDisplayName(fileName string) string {
	fileName = strings.TrimSuffix(fileName, ".pipe")
	fileName = strings.Replace(fileName, "::", " > ", -1)
	fileName = strings.Replace(fileName, "-", " ", -1)
	fileName = strings.Replace(fileName, "_", " ", -1)
	fileName = strings.Replace(fileName, ".", " ", -1)
	fileName = strings.Title(fileName)
	return fileName
}
