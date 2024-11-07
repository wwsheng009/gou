package process

import (
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/gou/runtime/v8/bridge"
	"rogchap.com/v8go"
)

// ExportFunction function template
func ExportFunction(iso *v8go.Isolate) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, exec)
}

// exec
func exec(info *v8go.FunctionCallbackInfo) *v8go.Value {

	jsArgs := info.Args()
	if len(jsArgs) < 1 {
		return bridge.JsException(info.Context(), "missing parameters")
	}

	if !jsArgs[0].IsString() {
		return bridge.JsException(info.Context(), "the first parameter should be a string")
	}

	share, err := bridge.ShareData(info.Context())
	if err != nil {
		return bridge.JsException(info.Context(), err)
	}

	goArgs := []interface{}{}
	if len(jsArgs) > 1 {
		for _, arg := range jsArgs[1:] {
			v, err := bridge.GoValue(arg, info.Context())
			if err != nil {
				return bridge.JsException(info.Context(), err)
			}
			goArgs = append(goArgs, v)
		}
	}
	process, err := process.Of(jsArgs[0].String(), goArgs...)
	if err != nil {
		return bridge.JsException(info.Context(), err)
	}
	goRes, err := process.
		WithGlobal(share.Global).
		WithSID(share.Sid).
		Exec()

	if err != nil {
		return bridge.JsException(info.Context(), err)
	}

	jsRes, err := bridge.JsValue(info.Context(), goRes)
	if err != nil {
		return bridge.JsException(info.Context(), err)
	}

	return jsRes
}
