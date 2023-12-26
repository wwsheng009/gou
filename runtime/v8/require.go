package v8

import (
	"fmt"

	"github.com/yaoapp/gou/runtime/v8/bridge"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/kun/log"
	"rogchap.com/v8go"
)

// Require function template
func Require(iso *v8go.Isolate) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {

		share, err := bridge.ShareData(info.Context())
		if err != nil {
			return bridge.JsException(info.Context(), err)
		}

		jsArgs := info.Args()
		if len(jsArgs) < 1 {
			return bridge.JsException(info.Context(), "missing parameters")
		}

		if !jsArgs[0].IsString() {
			return bridge.JsException(info.Context(), "the first parameter should be a string")
		}

		id := jsArgs[0].String()
		script := Scripts[id]
		if share.Root {
			if _, has := RootScripts[id]; has {
				script = RootScripts[id]
			}
		}

		globalName := "require"
		_, err := info.Context().RunScript(Transform(script.Source, globalName), script.File)
		if err != nil {
			message := fmt.Sprintf("failed to require file:%s, error:\n %+v.", script.File, err)
			log.Error("[V8] process error. %s", message)
			exception.New(fmt.Sprintf("%+v", err), 500).Throw()
		}

		global, _ := info.Context().Global().Get(globalName)
		return global
	})
}
