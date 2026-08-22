package gen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

type entityData struct {
	Pkg          string
	StructName   string
	TableName    string
	NeedsTime    bool
	NeedsBun     bool
	NeedsJSON    bool
	ExtraImports []entityImport // 自定义 import（来自 typeMapping/fieldMapping）
	Fields       []entityField
}

type entityImport struct {
	Alias string
	Path  string
}

type entityField struct {
	Name    string
	Type    string
	BunTag  string
	JsonTag string
	Comment string
}

func generateEntityCode(table string, columns []columnInfo, pkg, jsonCase string, typeMapping, fieldMapping map[string]TypeMappingItem, entityFieldEx string) string {
	data := entityData{
		Pkg:        pkg,
		StructName: ToPascalCase(table),
		TableName:  table,
		NeedsBun:   true,
	}

	extraImportSet := make(map[string]entityImport)
	excludedFields := parseEntityFieldEx(entityFieldEx)

	for _, c := range columns {
		if isFieldExcluded(table, c.Name, excludedFields) {
			continue
		}
		goType, importPath, importAlias := resolveGoType(table, c, typeMapping, fieldMapping)
		if goType == "time.Time" || goType == "*time.Time" {
			data.NeedsTime = true
		}
		if goType == "json.RawMessage" {
			data.NeedsJSON = true
		}
		if importPath != "" {
			// 同一路径重复出现时，优先保留带别名的导入配置。
			importItem := entityImport{Alias: importAlias, Path: importPath}
			if exists, ok := extraImportSet[importPath]; !ok || (exists.Alias == "" && importAlias != "") {
				extraImportSet[importPath] = importItem
			}
		}

		bunTag := buildBunTag(c)

		comment := ""
		if c.Comment != "" {
			comment = fmt.Sprintf(" // %s", c.Comment)
		}

		data.Fields = append(data.Fields, entityField{
			Name:    ToPascalCase(c.Name),
			Type:    goType,
			BunTag:  bunTag,
			JsonTag: toJsonCase(c.Name, jsonCase),
			Comment: comment,
		})
	}

	// 模板中已硬编码的 import，不再重复加入 ExtraImports
	builtinImports := map[string]struct{}{
		"encoding/json":          {},
		"time":                   {},
		"github.com/uptrace/bun": {},
	}
	importPaths := make([]string, 0, len(extraImportSet))
	for imp := range extraImportSet {
		importPaths = append(importPaths, imp)
	}
	sort.Strings(importPaths)
	for _, imp := range importPaths {
		if _, builtin := builtinImports[imp]; !builtin {
			data.ExtraImports = append(data.ExtraImports, extraImportSet[imp])
		}
	}

	var buf bytes.Buffer
	t := template.Must(template.New("entity").Parse(entityTpl))
	_ = t.Execute(&buf, data)
	return buf.String()
}

// resolveGoType 解析字段的 Go 类型，优先级：fieldMapping > typeMapping > 默认推导。
func resolveGoType(table string, c columnInfo, typeMapping, fieldMapping map[string]TypeMappingItem) (goType, importPath, importAlias string) {
	// 1. fieldMapping: 表名.字段名 精确匹配
	if fieldMapping != nil {
		key := table + "." + c.Name
		if m, ok := fieldMapping[key]; ok {
			return m.Type, m.Import, m.PkgAs
		}
	}
	// 2. typeMapping: 按数据库类型名匹配（同时检查 data_type 和 udt_name）
	if typeMapping != nil {
		for _, typeName := range []string{strings.ToLower(c.DataType), strings.ToLower(c.UdtName)} {
			if m, ok := typeMapping[typeName]; ok {
				return m.Type, m.Import, m.PkgAs
			}
		}
	}
	// 3. 默认推导
	return pgTypeToGo(c), "", ""
}

// buildBunTag 根据列元信息生成 bun struct tag 值。
func buildBunTag(c columnInfo) string {
	parts := []string{c.Name}

	// 主键（自增列直接标注 autoincrement，无需判断默认值）
	if c.IsPrimary {
		parts = append(parts, "pk")
		if c.AutoIncrement {
			parts = append(parts, "autoincrement")
		}
	}

	// SQL 类型标注
	if sqlType := bunSQLType(c); sqlType != "" {
		parts = append(parts, "type:"+sqlType)
	}

	// NOT NULL
	if c.IsNullable == "NO" && !c.IsPrimary {
		parts = append(parts, "notnull")
	}

	// unique（从约束查询更好，这里先跳过，用户可手动加）

	// nullzero: 时间类型、有默认值的非主键字段
	goType := pgTypeToGo(c)
	if goType == "time.Time" || goType == "*time.Time" {
		parts = append(parts, "nullzero")
	}

	// default
	if c.HasDefault && !c.IsPrimary {
		defVal := cleanDefault(c.DefaultValue)
		if defVal != "" {
			parts = append(parts, "default:"+defVal)
		}
	}

	// soft_delete: 名为 deleted_at 的时间列
	if c.Name == "deleted_at" && (goType == "time.Time" || goType == "*time.Time") {
		parts = append(parts, "soft_delete")
	}

	// array
	if strings.EqualFold(c.DataType, "array") || strings.EqualFold(c.DataType, "ARRAY") {
		parts = append(parts, "array")
	}

	return strings.Join(parts, ",")
}

// bunSQLType 返回需要显式标注的 SQL 类型，无需标注时返回空串。
func bunSQLType(c columnInfo) string {
	switch strings.ToLower(c.DataType) {
	case "uuid":
		return "uuid"
	case "jsonb":
		return "jsonb"
	case "json":
		return "json"
	case "numeric":
		if c.NumericPrecision > 0 {
			return fmt.Sprintf("numeric(%d,%d)", c.NumericPrecision, c.NumericScale)
		}
		return "numeric"
	case "character varying":
		if c.MaxLength > 0 {
			return fmt.Sprintf("varchar(%d)", c.MaxLength)
		}
		return ""
	case "inet":
		return "inet"
	case "cidr":
		return "cidr"
	case "macaddr":
		return "macaddr"
	case "interval":
		return "interval"
	default:
		return ""
	}
}

// cleanDefault 清理默认值表达式，去除类型转换后缀。
func cleanDefault(raw string) string {
	v := strings.TrimSpace(raw)
	// 跳过 nextval 序列（已在 autoincrement 处理）
	if strings.Contains(strings.ToLower(v), "nextval(") {
		return ""
	}
	// 去除 PostgreSQL 类型转换后缀，如 'value'::text
	if idx := strings.Index(v, "::"); idx >= 0 {
		v = v[:idx]
	}
	// 去除单引号包裹
	v = strings.Trim(v, "'")
	return v
}

// entityFieldExRule 表示一条字段排除规则。
type entityFieldExRule struct {
	Table string // 表名，"*" 表示匹配所有表
	Field string // 字段名
}

// parseEntityFieldEx 解析 entityFieldEx 配置字符串。
// 格式: "表名.字段名, *.字段名"，逗号分隔，支持 * 通配表名。
func parseEntityFieldEx(s string) []entityFieldExRule {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	rules := make([]entityFieldExRule, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.LastIndex(p, ".")
		if idx < 0 {
			continue
		}
		rules = append(rules, entityFieldExRule{
			Table: strings.TrimSpace(p[:idx]),
			Field: strings.TrimSpace(p[idx+1:]),
		})
	}
	return rules
}

// isFieldExcluded 判断指定表的字段是否应被排除。
func isFieldExcluded(table, field string, rules []entityFieldExRule) bool {
	for _, r := range rules {
		if r.Field != field {
			continue
		}
		if r.Table == "*" || r.Table == table {
			return true
		}
	}
	return false
}
