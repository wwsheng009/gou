package v8

import (
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/exception"
)

func init() {
	process.Register("scripts", processScripts)
	process.Register("studio", processStudio)
	process.Register("runtime.v8.stats", processV8IsoStats)
	process.Register("runtime.v8.statTotal", processV8TotalStat)
	process.Register("runtime.v8.option", processV8Option)
	process.Register("runtime.v8.restart", processV8Restart)
	process.Register("runtime.v8.stop", processV8Stop)
	process.Register("runtime.v8.Start", processV8Start)
}

// processScripts
func processScripts(process *process.Process) interface{} {

	script, err := Select(process.ID)
	if err != nil {
		exception.New("scripts.%s not loaded", 404, process.ID).Throw()
		return nil
	}

	return script.Exec(process)

	// script, err := Select(process.ID)
	// if err != nil {
	// 	exception.New("scripts.%s not loaded", 404, process.ID).Throw()
	// 	return nil
	// }

	// if runtimeOption.Mode == "normal" {
	// 	return runNormalMode(script, process.Sid, process.Global, process.Method, process.Args...)
	// }

	// ctx, err := script.NewContext(process.Sid, process.Global)
	// if err != nil {
	// 	message := fmt.Sprintf("scripts.%s failed to create context. %+v", process.ID, err)
	// 	log.Error("[V8] process error. %s", message)
	// 	exception.New(message, 500).Throw()
	// 	return nil
	// }
	// defer ctx.Close()

	// res, err := ctx.Call(process.Method, process.Args...)
	// if err != nil {
	// 	exception.New(err.Error(), 500).Throw()
	// }

	// return res
}

// wrk -t12 -c400 -d30s 'http://maxdev.yao.run/api/register/wechat/check/status?state=881119&sn=136-552-234'
// func runNormalMode(script *Script, sid string, data map[string]interface{}, method string, args ...interface{}) interface{} {

// 	// defer runtime.GC()
// 	// iso, err := SelectIso(2000 * time.Millisecond)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// defer iso.Unlock()

// 	iso := v8go.NewIsolate()
// 	defer iso.Dispose()

// 	tmpl := MakeTemplate(iso)
// 	ctx := v8go.NewContext(iso, tmpl)
// 	defer ctx.Close()

// 	// v, err := context.RunScript(script.Source, script.File)

// 	instance, err := iso.CompileUnboundScript(script.Source, script.File, v8go.CompileOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	v, err := instance.Run(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer v.Release()

// 	global := ctx.Global()
// 	jsArgs, err := bridge.JsValues(ctx, args)
// 	if err != nil {
// 		return fmt.Errorf("%s.%s %s", script.ID, method, err.Error())
// 	}

// 	defer bridge.FreeJsValues(jsArgs)

// 	goData := map[string]interface{}{
// 		"SID":  sid,
// 		"ROOT": script.Root,
// 		"DATA": data,
// 	}

// 	jsData, err := bridge.JsValue(ctx, goData)
// 	if err != nil {
// 		return err
// 	}

// 	err = global.Set("__yao_data", jsData)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		if !jsData.IsNull() && !jsData.IsUndefined() {
// 			jsData.Release()
// 		}
// 	}()

// 	jsRes, err := global.MethodCall(method, bridge.Valuers(jsArgs)...)
// 	if err != nil {
// 		return fmt.Errorf("%s.%s %+v", script.ID, method, err)
// 	}

// 	goRes, err := bridge.GoValue(jsRes, ctx)
// 	if err != nil {
// 		return fmt.Errorf("%s.%s %s", script.ID, method, err.Error())
// 	}

// 	return goRes

// }

// processScripts scripts.ID.Method
func processStudio(process *process.Process) interface{} {

	script, err := SelectRoot(process.ID)
	if err != nil {
		exception.New("studio.%s not loaded", 404, process.ID).Throw()
		return nil
	}
	return script.Exec(process)

}
func processV8TotalStat(process *process.Process) interface{} {
	total := HeapTotalStatistics{}
	isolates.Range(func(iso *Isolate) bool {
		stat := iso.HeapStat()
		total.TotalHeapSize += stat.TotalHeapSize
		total.TotalHeapSizeExecutable += stat.TotalHeapSizeExecutable
		total.TotalPhysicalSize += stat.TotalPhysicalSize
		total.TotalAvailableSize += stat.TotalAvailableSize
		total.UsedHeapSize += stat.UsedHeapSize
		total.HeapSizeLimit += stat.HeapSizeLimit
		total.MallocedMemory += stat.MallocedMemory
		total.ExternalMemory += stat.ExternalMemory
		total.PeakMallocedMemory += stat.PeakMallocedMemory
		total.NumberOfNativeContexts += stat.NumberOfNativeContexts
		total.NumberOfDetachedContexts += stat.NumberOfDetachedContexts
		total.Count += 1
		return true
	})
	total.Length = uint64(isolates.Len)

	return total
}
func processV8IsoStats(process *process.Process) interface{} {
	stats := make([]HeapStatistics, 0)
	isolates.Range(func(iso *Isolate) bool {
		stats = append(stats, iso.HeapStat())
		return true
	})
	return stats
}

func processV8Option(process *process.Process) interface{} {
	return runtimeOption
}

func processV8Stop(process *process.Process) interface{} {
	Stop()
	return map[string]interface{}{"status": "ok"}
}
func processV8Start(process *process.Process) interface{} {
	Start(runtimeOption)
	return map[string]interface{}{"status": "ok"}
}
func processV8Restart(process *process.Process) interface{} {
	Stop()

	Start(runtimeOption)
	return map[string]interface{}{"status": "ok"}
}
