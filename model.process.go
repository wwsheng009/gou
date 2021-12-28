package gou

import (
	"github.com/yaoapp/kun/any"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/kun/str"
)

// ModelHandlers 模型运行器
var ModelHandlers = map[string]ProcessHandler{
	"find":         processFind,
	"get":          processGet,
	"paginate":     processPaginate,
	"create":       processCreate,
	"update":       processUpdate,
	"save":         processSave,
	"delete":       processDelete,
	"destroy":      processDestroy,
	"insert":       processInsert,
	"updatewhere":  processUpdateWhere,
	"deletewhere":  processDeleteWhere,
	"destroywhere": processDestroyWhere,
	"eachsave":     processEachSave,
}

// processFind 运行模型 MustFind
func processFind(process *Process) interface{} {
	process.ValidateArgNums(2)
	mod := Select(process.Class)
	params, ok := process.Args[1].(QueryParam)
	if !ok {
		params = QueryParam{}
	}
	return mod.MustFind(process.Args[0], params)
}

// processGet 运行模型 MustGet
func processGet(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	params, ok := AnyToQueryParam(process.Args[0])
	if !ok {
		exception.New("第1个查询参数错误 %v", 400, process.Args[0]).Throw()
	}
	return mod.MustGet(params)
}

// processPaginate 运行模型 MustPaginate
func processPaginate(process *Process) interface{} {
	process.ValidateArgNums(3)
	mod := Select(process.Class)
	params, ok := AnyToQueryParam(process.Args[0])
	if !ok {
		exception.New("第1个查询参数错误 %v", 400, process.Args[0]).Throw()
	}

	page := any.Of(process.Args[1]).CInt()
	pagesize := any.Of(process.Args[2]).CInt()
	return mod.MustPaginate(params, page, pagesize)
}

// processCreate 运行模型 MustCreate
func processCreate(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	row := any.Of(process.Args[0]).Map().MapStrAny
	return mod.MustCreate(row)
}

// processUpdate 运行模型 MustUpdate
func processUpdate(process *Process) interface{} {
	process.ValidateArgNums(2)
	mod := Select(process.Class)
	id := process.Args[0]
	row := any.Of(process.Args[1]).Map().MapStrAny
	mod.MustUpdate(id, row)
	return nil
}

// processSave 运行模型 MustSave
func processSave(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	row := any.Of(process.Args[0]).Map().MapStrAny
	return mod.MustSave(row)
}

// processDelete 运行模型 MustDelete
func processDelete(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	mod.MustDelete(process.Args[0])
	return nil
}

// processDestroy 运行模型 MustDestroy
func processDestroy(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	mod.MustDestroy(process.Args[0])
	return nil
}

// processInsert 运行模型 MustInsert
func processInsert(process *Process) interface{} {
	process.ValidateArgNums(2)
	mod := Select(process.Class)
	var colums = []string{}
	colums, ok := process.Args[0].([]string)
	if !ok {
		anyColums, ok := process.Args[0].([]interface{})
		if !ok {
			exception.New("第1个查询参数错误 %v", 400, process.Args[0]).Throw()
		}
		for _, col := range anyColums {
			colums = append(colums, string(str.Of(col)))
		}
	}

	var rows = [][]interface{}{}
	rows, ok = process.Args[1].([][]interface{})
	if !ok {
		anyRows, ok := process.Args[1].([]interface{})
		if !ok {
			exception.New("第2个查询参数错误 %v", 400, process.Args[1]).Throw()
		}
		for _, anyRow := range anyRows {

			row, ok := anyRow.([]interface{})
			if !ok {
				exception.New("第2个查询参数错误 %v", 400, process.Args[1]).Throw()
			}
			rows = append(rows, row)
		}
	}

	mod.MustInsert(colums, rows)
	return nil
}

// processUpdateWhere 运行模型 MustUpdateWhere
func processUpdateWhere(process *Process) interface{} {
	process.ValidateArgNums(2)
	mod := Select(process.Class)
	params, ok := AnyToQueryParam(process.Args[0])
	if !ok {
		exception.New("第1个查询参数错误 %v", 400, process.Args[0]).Throw()
	}
	row := any.Of(process.Args[1]).Map().MapStrAny
	return mod.MustUpdateWhere(params, row)
}

// processDeleteWhere 运行模型 MustDeleteWhere
func processDeleteWhere(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	params, ok := AnyToQueryParam(process.Args[0])
	if !ok {
		params = QueryParam{}
	}
	return mod.MustDeleteWhere(params)
}

// processDestroyWhere 运行模型 MustDestroyWhere
func processDestroyWhere(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	params, ok := AnyToQueryParam(process.Args[0])
	if !ok {
		params = QueryParam{}
	}
	return mod.MustDestroyWhere(params)
}

// processEachSave 运行模型 MustEachSave
func processEachSave(process *Process) interface{} {
	process.ValidateArgNums(1)
	mod := Select(process.Class)
	rows := process.ArgsRecords(0)
	eachrow := map[string]interface{}{}
	if process.NumOfArgsIs(2) {
		eachrow = process.ArgsMap(1)
	}
	return mod.MustEachSave(rows, eachrow)
}