package gen

const serviceTpl = `// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package service
{{if .Imports}}
import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
)
{{end}}

type I{{.PascalName}} interface {
{{- range .Methods}}
	{{.Name}}({{.Params}}) {{.Results}}
{{- end}}
}

var local{{.PascalName}} I{{.PascalName}}

func {{.PascalName}}() I{{.PascalName}} {
	if local{{.PascalName}} == nil {
		panic("service {{.PascalName}} not registered")
	}
	return local{{.PascalName}}
}

func Register{{.PascalName}}(s I{{.PascalName}}) {
	local{{.PascalName}} = s
}
`

const logicTpl = `// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package logic

import (
{{- range .Packages}}
	_ "{{.}}"
{{- end}}
)
`
