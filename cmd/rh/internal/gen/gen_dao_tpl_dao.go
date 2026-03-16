package gen

const daoInternalTpl = `// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import "github.com/uptrace/bun"

const {{.PascalName}}Table = "{{.TableName}}"

// {{.ColumnsStructName}} defines the column names for table {{.TableName}}.
type {{.ColumnsStructName}} struct {
{{- range .Fields}}
	{{.Name}} string{{.Comment}}
{{- end}}
}

// {{.IdentsStructName}} defines the column ident for table {{.TableName}}.
type {{.IdentsStructName}} struct {
{{- range .Fields}}
	{{.Name}} bun.Ident{{.Comment}}
{{- end}}
}

var {{.PascalName}}ColumnsVar = {{.ColumnsStructName}}{
{{- range .Fields}}
	{{.Name}}: "{{.ColumnName}}",
{{- end}}
}

var {{.PascalName}}IdentsVar = {{.IdentsStructName}}{
{{- range .Fields}}
	{{.Name}}: bun.Ident("{{.ColumnName}}"),
{{- end}}
}
`

const tableTpl = `// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package {{.Pkg}}

import (
	"{{.EntityImportPath}}"
	"{{.TableImportPath}}/internal"
)

var {{.PascalName}} = &{{.CamelName}}Table{}

type {{.CamelName}}Table struct{}

func (*{{.CamelName}}Table) Model() *ent.{{.PascalName}} {
	return (*ent.{{.PascalName}})(nil)
}

func (*{{.CamelName}}Table) Table() string {
	return internal.{{.PascalName}}Table
}

func (*{{.CamelName}}Table) Columns() internal.{{.PascalName}}Columns {
	return internal.{{.PascalName}}ColumnsVar
}

func (*{{.CamelName}}Table) Idents() internal.{{.PascalName}}Idents {
	return internal.{{.PascalName}}IdentsVar
}
`
