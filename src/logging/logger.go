package logging

import (
	"fmt"
	customstrings "github.com/Layer9Berlin/pipedream/src/custom/strings"
	"github.com/logrusorgru/aurora/v3"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"sort"
	"strings"
)

// UserPipeLogLevel is the level at which user-defined pipes will be logged
var UserPipeLogLevel = logrus.InfoLevel

// BuiltInPipeLogLevel is the level at which built-in pipes will be logged
var BuiltInPipeLogLevel = logrus.ErrorLevel

// CustomFormatter prints log entries in a format suitable for output to the console
type CustomFormatter struct {
}

// Format turns a log entry into a string
func (formatter CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if readerField, ok := entry.Data["reader"]; ok {
		if reader, ok := readerField.(io.Reader); ok {
			result, err := ioutil.ReadAll(reader)
			return result, err
		}
	}

	indentation, ok := entry.Data["indentation"].(int)
	if !ok {
		indentation = 0
	}

	result := coloredOutput(entry,
		fmt.Sprint(
			strings.Repeat(" ", indentation),
			extractField(entry, "prefix"),
			extractFields(entry, "middleware", "message", "info"),
		),
	) + "\n"

	return []byte(result), nil
}

func coloredOutput(entry *logrus.Entry, subject interface{}) string {
	colorOverride := extractField(entry, "color")
	if colorOverride != "" {
		switch colorOverride {
		case "red":
			return fmt.Sprint(aurora.Red(subject))
		case "yellow":
			return fmt.Sprint(aurora.Yellow(subject))
		case "blue":
			return fmt.Sprint(aurora.Blue(subject))
		case "cyan":
			return fmt.Sprint(aurora.Cyan(subject))
		case "black":
			return fmt.Sprint(aurora.Black(subject))
		case "lightgray", "lightgrey":
			return fmt.Sprint(aurora.Gray(18, subject))
		case "gray", "grey":
			return fmt.Sprint(aurora.Gray(12, subject))
		case "green":
			return fmt.Sprint(aurora.Green(subject))
		}
	}
	switch entry.Level {
	case logrus.ErrorLevel:
		return fmt.Sprint(aurora.Red(subject))
	case logrus.WarnLevel:
		return fmt.Sprint(aurora.Yellow(subject))
	case logrus.InfoLevel:
		return fmt.Sprint(aurora.Blue(subject))
	case logrus.DebugLevel:
		return fmt.Sprint(aurora.Gray(12, subject))
	case logrus.TraceLevel:
		return fmt.Sprint(aurora.Gray(18, subject))
	}
	return fmt.Sprint(subject)
}

func extractField(entry *logrus.Entry, key string) string {
	result, ok := entry.Data[key]
	if ok {
		maxLength := 128
		if entry.Level >= logrus.DebugLevel {
			maxLength = 1024
		}
		return prettyPrint(result, maxLength)
	}
	return ""
}

func extractFields(entry *logrus.Entry, keys ...string) string {
	fields := make([]string, 0, len(keys))
	for _, key := range keys {
		result, ok := entry.Data[key]
		if ok {
			maxLength := 128
			if entry.Level <= logrus.ErrorLevel {
				maxLength = 1024
			}
			fields = append(fields, prettyPrint(result, maxLength))
		}
	}
	return strings.Join(fields, " | ")
}

func prettyPrint(info interface{}, maxLength int) string {
	infoMap, ok := info.(map[string]interface{})
	if ok {
		return prettyPrintMap(infoMap, maxLength)
	}
	infoArray, ok := info.([]string)
	if ok {
		return prettyPrintArray(infoArray, maxLength)
	}
	infoString := ""
	if info != nil {
		infoString = fmt.Sprint(info)
	}
	return customstrings.Shorten(infoString, maxLength)
}

func prettyPrintArray(arrayToPrint []string, maxLength int) string {
	for index, item := range arrayToPrint {
		arrayToPrint[index] = customstrings.Shorten(item, maxLength)
	}
	return customstrings.Shorten(strings.Join(arrayToPrint, ", "), maxLength)
}

func prettyPrintMap(mapToPrint map[string]interface{}, maxLength int) string {
	if len(mapToPrint) > 0 {
		keys := make([]string, 0, len(mapToPrint))
		for k := range mapToPrint {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		result := "{ "
		for _, key := range keys {
			stringValue, ok := mapToPrint[key].(string)
			if ok {
				if len(result) > 2 {
					result = result + ", "
				}
				maxComponentLength := int(math.Floor(float64(maxLength-4) / 2))
				result = result + fmt.Sprintf("%v: `%v`", customstrings.Shorten(key, maxComponentLength), customstrings.Shorten(stringValue, maxComponentLength))
			}
		}
		result = result + " }"
		return result
	}
	return ""
}
