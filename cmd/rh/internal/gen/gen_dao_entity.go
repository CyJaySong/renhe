package gen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type entityData struct {
	Pkg        string
	StructName string
	TableName  string
	NeedsTime  bool
	NeedsBun   bool
	NeedsJSON  bool
	Fields     []entityField
}

type entityField struct {
	Name    string
	Type    string
	BunTag  string
	JsonTag string
	Comment string
}

func generateEntityCode(table string, columns []columnInfo, pkg, jsonCase string) string {
	data := entityData{
		Pkg:        pkg,
		StructName: ToPascalCase(table),
		TableName:  table,
		NeedsBun:   true,
	}

	for _, c := range columns {
		goType := pgTypeToGo(c)
		if goType == "time.Time" || goType == "*time.Time" {
			data.NeedsTime = true
		}
		if goType == "json.RawMessage" {
			data.NeedsJSON = true
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

	var buf bytes.Buffer
	t := template.Must(template.New("entity").Parse(entityTpl))
	_ = t.Execute(&buf, data)
	return buf.String()
}

// buildBunTag 根据列元信息生成 bun struct tag 值。
func buildBunTag(c columnInfo) string {
	parts := []string{c.Name}

	// 主键
	if c.IsPrimary {
		parts = append(parts, "pk")
		if c.HasDefault && isAutoIncrement(c.DefaultValue) {
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

// isAutoIncrement 判断默认值是否为自增序列。
func isAutoIncrement(defVal string) bool {
	d := strings.ToLower(defVal)
	return strings.Contains(d, "nextval(") || strings.Contains(d, "generated")
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
