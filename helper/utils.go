package helper

import (
	"fmt"

	"github.com/fatih/color"
	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/runtime/v8/bridge"
)

// Dump The Dump function dumps the given variables:
func Dump(values ...interface{}) {

	f := NewFormatter()
	f.Indent = 4
	f.RawStrings = true
	for _, v := range values {

		if err, ok := v.(error); ok {
			color.Red(err.Error())
			continue
		}

		switch value := v.(type) {

		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
			color.Cyan(fmt.Sprintf("%v", v))
			continue

		case string, []byte:
			color.Green(fmt.Sprintf("%s", v))
			continue

		case bridge.UndefinedT:
			color.Magenta(value.String())
			continue

		case bridge.FunctionT:
			color.Cyan(value.String())
			continue

		case bridge.PromiseT:
			color.Cyan("Promise { " + value.String() + " }")
			continue

		default:
			var res interface{}
			txt, err := jsoniter.Marshal(v)
			if err != nil {
				color.Red(err.Error())
				continue
			}

			jsoniter.Unmarshal(txt, &res)
			bytes, _ := f.Marshal(res)
			fmt.Println(string(bytes))
		}
	}
}

func PrintString(values ...interface{}) {

	f := NewFormatter()
	f.Indent = 4
	f.RawStrings = true
	for _, v := range values {

		if err, ok := v.(error); ok {
			color.Red(err.Error())
			continue
		}

		switch value := v.(type) {

		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
			fmt.Print(color.CyanString(fmt.Sprintf("%v", v)))
			continue

		case string, []byte:
			fmt.Print(color.GreenString(fmt.Sprintf("%s", v)))
			continue

		case bridge.UndefinedT:
			fmt.Print(color.MagentaString(value.String()))
			continue

		case bridge.FunctionT:
			fmt.Print(color.CyanString(value.String()))
			continue

		case bridge.PromiseT:
			fmt.Print(color.CyanString("Promise { " + value.String() + " }"))
			continue

		default:
			var res interface{}
			txt, err := jsoniter.Marshal(v)
			if err != nil {
				color.Red(err.Error())
				continue
			}

			jsoniter.Unmarshal(txt, &res)
			bytes, _ := f.Marshal(res)
			fmt.Print(string(bytes))
		}
	}
}
