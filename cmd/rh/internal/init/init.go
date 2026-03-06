package initcmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func Run(name string, args []string) error {
	var root string
	if len(args) >= 2 {
		root, _ = filepath.Abs(args[1])
	} else {
		base, _ := os.Getwd()
		root = filepath.Join(base, name)
	}

	dirs := []string{
		"api",
		"hack",
		"internal/cmd",
		"internal/consts",
		"internal/controller",
		"internal/logic",
		"internal/model/do",
		"internal/model/ent",
		"internal/model/table",
		"internal/service",
		"manifest/config",
		"manifest/docker",
		"manifest/deploy",
		"resource",
		"utility",
	}

	for _, d := range dirs {
		p := filepath.Join(root, d)
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", p, err)
		}
		fmt.Printf("  created: %s/\n", d)
	}

	data := tplData{
		Module:          name,
		FrameworkModule: frameworkModule,
	}

	files := map[string]string{
		"main.go":                     mainGoTpl,
		"internal/cmd/cmd.go":         cmdGoTpl,
		"manifest/config/config.yaml": configYamlTpl,
		"hack/config.yaml":            hackConfigYamlTpl,
		"go.mod":                      goModTpl,
		".gitignore":                  gitignoreTpl,
	}

	for f, tplStr := range files {
		content, err := renderTpl(f, tplStr, data)
		if err != nil {
			return fmt.Errorf("failed to render %s: %w", f, err)
		}
		p := filepath.Join(root, f)
		if err := os.WriteFile(p, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", p, err)
		}
		fmt.Printf("  created: %s\n", f)
	}

	fmt.Printf("\nProject '%s' initialized successfully!\n", name)
	fmt.Printf("  cd %s && go mod tidy\n", root)
	return nil
}

func renderTpl(name, tplStr string, data tplData) ([]byte, error) {
	t, err := template.New(name).Parse(tplStr)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
