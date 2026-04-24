package gen

const entityTpl = `// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package {{.Pkg}}

import (
{{- if .NeedsJSON}}
	"encoding/json"
{{- end}}
{{- if .NeedsTime}}
	"time"
{{- end}}

	"github.com/uptrace/bun"
{{- range .ExtraImports}}
{{- if .Alias}}
	{{.Alias}} "{{.Path}}"
{{- else}}
	"{{.Path}}"
{{- end}}
{{- end}}
)

// {{.StructName}} is the entity for table {{.TableName}}.
type {{.StructName}} struct {
	bun.BaseModel ` + "`" + `bun:"table:{{.TableName}}"` + "`" + `

{{- range .Fields}}
	{{.Name}} {{.Type}} ` + "`" + `bun:"{{.BunTag}}" json:"{{.JsonTag}}"` + "`" + `{{.Comment}}
{{- end}}
}
`
