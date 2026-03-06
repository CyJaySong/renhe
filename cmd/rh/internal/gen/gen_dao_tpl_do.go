package gen

const doTpl = `// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package {{.Pkg}}

// {{.StructName}} is the data operation struct for table {{.TableName}}.
type {{.StructName}} struct {
{{- range .Fields}}
	{{.Name}} any{{.Comment}}
{{- end}}
}
`
