package gen

const entityTpl = `// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package {{.Pkg}}
{{if or .NeedsTime .NeedsBun}}
import (
{{- if .NeedsTime}}
	"time"
{{- end}}

	"github.com/uptrace/bun"
)
{{end}}

// {{.StructName}} is the entity for table {{.TableName}}.
type {{.StructName}} struct {
	bun.BaseModel ` + "`" + `bun:"table:{{.TableName}}"` + "`" + `

{{- range .Fields}}
	{{.Name}} {{.Type}} ` + "`" + `bun:"{{.BunTag}}" json:"{{.JsonTag}}"` + "`" + `{{.Comment}}
{{- end}}
}
`
