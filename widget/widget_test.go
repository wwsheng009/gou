package widget

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/yaoapp/gou/runtime"
	"github.com/yaoapp/kun/utils"
)

func TestLoad(t *testing.T) {
	w := load(t)
	v, err := w.ScriptExec("helper", "Foo", "Hello")
	if err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, "dyform", w.Name)
	assert.Equal(t, "Dynamic Form", w.Label)
	assert.Equal(t, "A form widget. users can design forms online", w.Description)
	assert.Equal(t, "0.1.0", w.Version)
	assert.Equal(t, "Hello World", v)
}

func TestInstanceLoad(t *testing.T) {
	w := load(t)
	err := w.Load()
	if err != nil {
		t.Fatal(err)
	}
}

func load(t *testing.T) *Widget {
	root := os.Getenv("GOU_TEST_APP_ROOT")
	path := filepath.Join(root, "widgets", "dyform")
	widget, err := Load(path, yao())
	if err != nil {
		t.Fatal(err)
	}
	return widget
}

func yao() *runtime.Runtime {
	return runtime.Yao(1024).
		AddFunction("UnitTestFn", func(global map[string]interface{}, sid string, args ...interface{}) interface{} {
			utils.Dump(global, sid, args)
			return args
		}).
		AddFunction("Process", func(global map[string]interface{}, sid string, args ...interface{}) interface{} {
			return map[string]interface{}{"global": global, "sid": sid, "args": args}
		}).
		AddObject("console", map[string]func(global map[string]interface{}, sid string, args ...interface{}) interface{}{
			"log": func(_ map[string]interface{}, _ string, args ...interface{}) interface{} {
				utils.Dump(args)
				return nil
			},
		})
}
