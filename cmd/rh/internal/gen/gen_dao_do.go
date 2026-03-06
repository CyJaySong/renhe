package gen

import (
	"bytes"
	"fmt"
	"text/template"
)

type doData struct {
	Pkg        string
	StructName string
	TableName  string
	Fields     []doField
}

type doField struct {
	Name    string
	Comment string
}

func generateDoCode(table string, columns []columnInfo, pkg string) string {
	data := doData{
		Pkg:        pkg,
		StructName: ToPascalCase(table),
		TableName:  table,
	}

	for _, c := range columns {
		comment := ""
		if c.Comment != "" {
			comment = fmt.Sprintf(" // %s", c.Comment)
		}
		data.Fields = append(data.Fields, doField{
			Name:    ToPascalCase(c.Name),
			Comment: comment,
		})
	}

	var buf bytes.Buffer
	t := template.Must(template.New("do").Parse(doTpl))
	_ = t.Execute(&buf, data)
	return buf.String()
}
