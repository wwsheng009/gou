package model

import (
	"strings"

	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/kun/maps"
	"github.com/yaoapp/xun"
)

// RunV2 执行查询栈
func (stack *QueryStack) RunV2() []maps.MapStrAny {
	res := [][]maps.MapStrAny{}
	indexMap := map[string]int{"": 0} // 记录 ExportPrefix 对应的 res 索引
	for i, qb := range stack.Builders {
		param := stack.Params[i]
		// 找回父级数据集
		parentRows := stack.getParentRows(i, res, indexMap)
		switch param.Relation.Type {
		case "hasMany":
			stack.runHasManyV2(&res, qb, param, parentRows)
		default:
			stack.runV2(&res, qb, param, parentRows)
		}

		// 记录当前结果索引，供子级引用
		indexMap[param.ExportPrefix] = len(res) - 1
	}

	if len(res) == 0 {
		return nil
	}
	return res[0]
}

// PaginateV2 执行查询栈(分页查询)
func (stack *QueryStack) PaginateV2(page int, pagesize int) maps.MapStrAny {
	res := [][]maps.MapStrAny{}
	indexMap := map[string]int{"": 0}
	var pageInfo xun.P
	for i, qb := range stack.Builders {
		param := stack.Params[i]
		if i == 0 {
			pageInfo = stack.paginate(page, pagesize, &res, qb, param)
			indexMap[param.ExportPrefix] = 0
			continue
		}
		// 找回父级数据集
		parentRows := stack.getParentRows(i, res, indexMap)
		switch param.Relation.Type {
		case "hasMany":
			stack.runHasManyV2(&res, qb, param, parentRows)
		default:
			stack.runV2(&res, qb, param, parentRows)
		}
		indexMap[param.ExportPrefix] = len(res) - 1
	}

	if len(res) == 0 {
		return nil
	}

	response := maps.MapStrAny{}
	response["data"] = res[0]
	response["pagesize"] = pageInfo.PageSize
	response["pagecnt"] = pageInfo.TotalPages
	response["pagesize"] = pageInfo.PageSize
	response["page"] = pageInfo.CurrentPage
	response["next"] = pageInfo.NextPage
	response["prev"] = pageInfo.PreviousPage
	response["total"] = pageInfo.Total
	return response
}

// runV2 执行 hasOne / belongsTo 查询并绑定数据
func (stack *QueryStack) runV2(res *[][]maps.MapStrAny, builder QueryStackBuilder, param QueryStackParam, parentRows []maps.MapStrAny) {

	// 1. 获取关系定义
	rel := param.Relation
	foreignIDs := []interface{}{}

	// 修复核心：从传入的确定的父级数据集中提取关联 ID
	for _, row := range parentRows {
		id := row.Get(rel.Foreign)
		if id != nil {
			foreignIDs = append(foreignIDs, id)
		}
	}

	// 2. 修正调试日志：直接访问 builder.Query 接口
	limit := 100
	if param.QueryParam.Limit > 0 {
		limit = param.QueryParam.Limit
	}

	// 3. 处理查询名称与空 ID 情况
	name := rel.Key
	if param.QueryParam.Alias != "" {
		name = param.QueryParam.Alias + "." + name
	}

	// if len(foreignIDs) == 0 {
	// 	*res = append(*res, []maps.MapStrAny{})
	// 	return
	// }
	builder.Query.Limit(limit)
	if len(foreignIDs) > 0 {
		builder.Query.WhereIn(name, foreignIDs)
	}
	// 4. 执行 WhereIn 查询
	// builder.Query.WhereIn(name, foreignIDs).Limit(limit)
	if param.QueryParam.Offset > 0 {
		builder.Query.Offset(param.QueryParam.Offset)
	}
	if param.QueryParam.Debug {
		defer log.With(log.F{
			"sql":      builder.Query.ToSQL(),
			"bindings": builder.Query.GetBindings()}).
			Trace("QueryStack run()")
	}
	rows := builder.Query.MustGet()

	// 5. 格式化数据并建立单条记录映射 (Map 结构)
	fmtRowMap := map[interface{}]maps.MapStrAny{}
	fmtRows := []maps.MapStrAny{}

	for _, row := range rows {
		fmtRow := maps.MapStrAny{}
		for key, value := range row {
			if cmap, has := builder.ColumnMap[key]; has {
				fmtRow[cmap.Export] = value
				cmap.Column.FliterOut(value, fmtRow, cmap.Export)
				continue
			}
			fmtRow[key] = value
		}
		unDotRow := fmtRow.UnDot()
		fmtRows = append(fmtRows, unDotRow)
		relVal := fmtRow.Get(rel.Key)
		if relVal != nil {
			fmtRowMap[relVal] = unDotRow // hasOne/belongsTo 只取最后匹配的一条
		}
	}

	// 6. 将结果绑定到父级数据集 (parentRows)
	varname := rel.Name
	for idx, prow := range parentRows {
		id := prow.Get(rel.Foreign)
		if row, has := fmtRowMap[id]; has {
			parentRows[idx][varname] = row
		} else {
			parentRows[idx][varname] = nil // 未匹配时显式设为 nil
		}
	}

	// 7. 追加结果到栈，供可能存在的嵌套 withs 使用
	*res = append(*res, fmtRows)
	stack.Next()
}

// runHasManyV2 执行 hasMany 查询并绑定数据
func (stack *QueryStack) runHasManyV2(res *[][]maps.MapStrAny, builder QueryStackBuilder, param QueryStackParam, parentRows []maps.MapStrAny) {

	// 获取上次查询结果，拼接结果集ID
	rel := stack.Relation()
	foreignIDs := []interface{}{}
	// 修复核心：使用传入的 parentRows 提取结果集 ID，而不是从 res 栈顶获取
	for _, row := range parentRows {
		id := row.Get(rel.Foreign)
		if id != nil {
			foreignIDs = append(foreignIDs, id)
		}
	}

	// 3. 处理查询字段名（处理 Alias）
	name := rel.Key
	if param.QueryParam.Alias != "" {
		name = param.QueryParam.Alias + "." + name
	}

	// 4. 处理空数据情况：确保父集合对应字段初始化为空切片
	if len(foreignIDs) == 0 {
		*res = append(*res, []maps.MapStrAny{})
		varname := rel.Name
		for idx := range parentRows {
			parentRows[idx][varname] = []maps.MapStrAny{}
		}
		return
	}

	// 5. 执行查询逻辑：修正 builder.Query 的调用方式
	limit := 100
	if param.QueryParam.Limit > 0 {
		limit = param.QueryParam.Limit
	}
	if len(foreignIDs) > 0 {
		builder.Query.WhereIn(name, foreignIDs)
	}
	builder.Query.Limit(limit)
	if param.QueryParam.Offset > 0 {
		builder.Query.Offset(param.QueryParam.Offset)
	}
	if param.QueryParam.Debug {
		defer log.With(log.F{
			"sql":      builder.Query.ToSQL(),
			"bindings": builder.Query.GetBindings()}).
			Trace("QueryStack runHasMany()")
	}
	// 获取子表数据
	rows := builder.Query.MustGet()

	// 6. 格式化数据并建立映射映射
	fmtRowMap := map[interface{}][]maps.MapStrAny{}
	fmtRows := []maps.MapStrAny{}

	for _, row := range rows {
		fmtRow := maps.MapStrAny{}
		for key, value := range row {
			if cmap, has := builder.ColumnMap[key]; has {
				fmtRow[cmap.Export] = value
				cmap.Column.FliterOut(value, fmtRow, cmap.Export)
				continue
			}
			fmtRow[key] = value
		}

		// 获取关联 Key 的值，用于后续分组映射
		relVal := fmtRow.Get(rel.Key)
		if relVal != nil {
			unDotRow := fmtRow.UnDot() // 处理多级路径
			fmtRows = append(fmtRows, unDotRow)

			if _, has := fmtRowMap[relVal]; !has {
				fmtRowMap[relVal] = []maps.MapStrAny{}
			}
			fmtRowMap[relVal] = append(fmtRowMap[relVal], unDotRow)
		}
	}

	// 7. 将结果绑定回父级数据集 (parentRows)
	varname := rel.Name
	for idx, prow := range parentRows {
		id := prow.Get(rel.Foreign)

		// 如果在 fmtRowMap 中找到了匹配的子表行，则绑定
		if subRows, has := fmtRowMap[id]; has {
			if _, initialized := parentRows[idx][varname]; !initialized {
				parentRows[idx][varname] = []maps.MapStrAny{}
			}
			// 追加数据
			parentRows[idx][varname] = append(parentRows[idx][varname].([]maps.MapStrAny), subRows...)
		} else {
			// 如果没找到，初始化为空切片以保证输出结构一致
			if _, initialized := parentRows[idx][varname]; !initialized {
				parentRows[idx][varname] = []maps.MapStrAny{}
			}
		}
	}

	// 8. 将本次查询的结果压入结果栈，供后续更深层（嵌套）的 withs 使用
	*res = append(*res, fmtRows)
	stack.Next()
}
func (stack *QueryStack) getParentRows(i int, res [][]maps.MapStrAny, indexMap map[string]int) []maps.MapStrAny {
	if i == 0 || len(res) == 0 {
		return nil
	}

	param := stack.Params[i]
	prefix := param.ExportPrefix

	// 如果没有点号，说明是第一层关联，父级必然是主表 res[0]
	if prefix == "" || !strings.Contains(strings.TrimRight(prefix, "."), ".") {
		return res[0]
	}

	// 处理嵌套关联：从 "rel.sub." 找到 "rel."
	parts := strings.Split(strings.TrimRight(prefix, "."), ".")
	if len(parts) > 1 {
		parentPath := strings.Join(parts[:len(parts)-1], ".") + "."
		if idx, ok := indexMap[parentPath]; ok {
			return res[idx]
		}
	}

	return res[0]
}
