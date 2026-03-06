package gen

import (
	"bytes"
	"fmt"
	"text/template"
)

type entityData struct {
	Pkg        string
	StructName string
	TableName  string
	NeedsTime  bool
	NeedsBun   bool
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
		goType := pgTypeToGo(c.DataType, c.UdtName, c.IsNullable == "YES")
		if goType == "time.Time" || goType == "*time.Time" {
			data.NeedsTime = true
		}

		bunTag := c.Name
		if c.IsPrimary {
			bunTag += ",pk"
		}
		if c.HasDefault && c.IsPrimary {
			bunTag += ",autoincrement"
		}

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
