package assert

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"reflect"
	"runtime/debug"
	"time"
)

var assertData map[string]any = map[string]any{}
var writer io.Writer

func AddAssertData(key string, value any) {
	assertData[key] = value
}

func RemoveAssertData(key string) {
	delete(assertData, key)
}

func ToWriter(w io.Writer) {
	writer = w
}

func stringify(item any) string {
	if item == nil {
		return "nil"
	}

	switch t := item.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", item)
	default:
		d, err := json.Marshal(t)
		if err != nil {
			return string(d)
		}
	}
	return fmt.Sprintf("%s", item)
}

func runAssert(msg string, args ...any) {
	if writer != nil {
		fmt.Fprintln(writer, GetTime())
		fmt.Fprintln(writer)
	} else {
		fmt.Println(GetTime())
		fmt.Println()
	}
	
	for k, v := range assertData {
		if writer != nil {
			fmt.Fprintf(writer, "%s=%s\n", k, stringify(v))
		} else {
			slog.Error("context", "key", k, "value", stringify(v))
		}
	}

	if writer != nil {
		fmt.Fprintf(writer, "%s: ", msg)
	} else {
		fmt.Printf("%s: ", msg)
	}

	for _, item := range args {
		if writer != nil {
			fmt.Fprintf(writer, "%s: ", stringify(item))
		} else {
			fmt.Printf("%v ", stringify(item))
		}
	}


	if writer != nil {
		fmt.Fprintln(writer)
		writer.Write(debug.Stack())
	} else {
		fmt.Println()
		debug.PrintStack()
	}

	if writer == nil {
		log.Fatal("runtime assert failure")
	} else {
		log.Fatal("runtime assert, dumped to provided file")
	}
}

// TODO: Think about passing around a context for debugging purposes
func Assert(truth bool, msg string, data ...any) {
	if !truth {
		runAssert(msg, data...)
	}
}

func NotNil(item any, msg string) {
	if item == nil || (reflect.ValueOf(item).Kind() == reflect.Ptr && reflect.ValueOf(item).IsNil()) {
		slog.Error("NotNil#nil encountered")
		runAssert(msg)
	}
}

func Never(msg string, data ...any) {
    Assert(false, msg, data...)
}
func NoError(err error, msg string, data ...any) {
	if err != nil {
		slog.Error("NoError#error encountered", "error", err)
		runAssert(msg, data...)
	}
}

func GetTime() string {
	return time.Now().Format("01-02-2006 15:04:05")
}