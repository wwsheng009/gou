package v8

import (
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/exception"
)

func init() {
	process.Register("scripts", processScripts)
	process.Register("studio", processStudio)
	// process.Register("runtime.v8.stats", processV8IsoStats)
	// process.Register("runtime.v8.statTotal", processV8TotalStat)
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
}

// processScripts scripts.ID.Method
func processStudio(process *process.Process) interface{} {

	script, err := SelectRoot(process.ID)
	if err != nil {
		exception.New("studio.%s not loaded", 404, process.ID).Throw()
		return nil
	}
	return script.Exec(process)

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
