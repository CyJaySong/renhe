package gen

import (
	"bytes"
	"fmt"
	"text/template"
)

type daoInternalData struct {
	TableName         string
	PascalName        string
	ColumnsStructName string
	IdentsStructName  string
	Fields            []daoInternalField
}

type daoInternalField struct {
	Name       string
	ColumnName string
	Comment    string
}

type tableFileData struct {
	Pkg             string
	PascalName      string
	CamelName       string
	TableImportPath string
}

func generateDaoInternal(table string, cols []columnInfo) string {
	pascal := ToPascalCase(table)
	data := daoInternalData{
		TableName:         table,
		PascalName:        pascal,
		ColumnsStructName: pascal + "Columns",
		IdentsStructName:  pascal + "Idents",
	}

	for _, c := range cols {
		comment := ""
		if c.Comment != "" {
			comment = fmt.Sprintf(" // %s", c.Comment)
		}
		data.Fields = append(data.Fields, daoInternalField{
			Name:       ToPascalCase(c.Name),
			ColumnName: c.Name,
			Comment:    comment,
		})
	}

	var buf bytes.Buffer
	t := template.Must(template.New("daoInternal").Parse(daoInternalTpl))
	_ = t.Execute(&buf, data)
	return buf.String()
}

func generateTableFile(table, pkg, tableImportPath string) string {
	pascal := ToPascalCase(table)
	data := tableFileData{
		Pkg:             pkg,
		PascalName:      pascal,
		CamelName:       toCamelCase(pascal),
		TableImportPath: tableImportPath,
	}

	var buf bytes.Buffer
	t := template.Must(template.New("table").Parse(tableTpl))
	_ = t.Execute(&buf, data)
	return buf.String()
}
