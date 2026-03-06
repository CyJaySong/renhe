package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ServiceConfig struct {
	SrcPath string
	DstPath string
	Module  string
}

type servicePackageInfo struct {
	PkgName    string
	PascalName string
	Methods    []serviceMethod
	Imports    []string
}

type serviceMethod struct {
	Name    string
	Params  string
	Results string
}

func RunService(cfg ServiceConfig) error {
	entries, err := os.ReadDir(cfg.SrcPath)
	if err != nil {
		return fmt.Errorf("failed to read logic directory %s: %w", cfg.SrcPath, err)
	}

	if err := os.MkdirAll(cfg.DstPath, 0755); err != nil {
		return fmt.Errorf("failed to create service directory %s: %w", cfg.DstPath, err)
	}

	var count int
	var logicPkgPaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgDir := filepath.Join(cfg.SrcPath, entry.Name())
		info, err := parseLogicPackage(pkgDir, entry.Name(), cfg.Module)
		if err != nil {
			return fmt.Errorf("failed to parse logic package %s: %w", entry.Name(), err)
		}
		if len(info.Methods) == 0 {
			continue
		}

		code, err := renderServiceFile(info)
		if err != nil {
			return fmt.Errorf("failed to render service for %s: %w", entry.Name(), err)
		}

		outFile := filepath.Join(cfg.DstPath, entry.Name()+".go")
		if err := writeFormattedService(outFile, code); err != nil {
			return fmt.Errorf("failed to write %s: %w", outFile, err)
		}
		fmt.Printf("  generated: %s (service)\n", outFile)
		logicPkgPaths = append(logicPkgPaths, cfg.Module+"/"+filepath.ToSlash(pkgDir))
		count++
	}

	if count == 0 {
		fmt.Println("No logic packages found to generate service interfaces.")
	} else {
		if err := generateLogicEntry(cfg.SrcPath, logicPkgPaths); err != nil {
			return fmt.Errorf("failed to generate logic entry: %w", err)
		}
		fmt.Printf("\nService generation completed! %d package(s) processed.\n", count)
	}
	return nil
}

func parseLogicPackage(dir, pkgName, module string) (*servicePackageInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	info := &servicePackageInfo{
		PkgName:    pkgName,
		PascalName: ToPascalCase(pkgName),
	}

	importSet := make(map[string]bool)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			fileImports := collectFileImports(file)
			for _, method := range extractMethods(fset, file, fileImports, importSet) {
				info.Methods = append(info.Methods, method)
			}
		}
	}

	for imp := range importSet {
		info.Imports = append(info.Imports, imp)
	}

	return info, nil
}

func collectFileImports(file *ast.File) map[string]string {
	imports := make(map[string]string)
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		var name string
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}
		imports[name] = path
	}
	return imports
}

func extractMethods(fset *token.FileSet, file *ast.File, fileImports map[string]string, importSet map[string]bool) []serviceMethod {
	var methods []serviceMethod

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || !fn.Name.IsExported() {
			continue
		}

		params := formatFieldList(fset, fn.Type.Params, fileImports, importSet)
		results := formatFieldList(fset, fn.Type.Results, fileImports, importSet)

		resultStr := results
		if fn.Type.Results != nil && len(fn.Type.Results.List) > 1 {
			resultStr = "(" + results + ")"
		}

		methods = append(methods, serviceMethod{
			Name:    fn.Name.Name,
			Params:  params,
			Results: resultStr,
		})
	}
	return methods
}

func formatFieldList(fset *token.FileSet, fl *ast.FieldList, fileImports map[string]string, importSet map[string]bool) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}

	var parts []string
	for _, field := range fl.List {
		typeStr := exprToString(fset, field.Type)
		collectImportsFromExpr(field.Type, fileImports, importSet)

		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			for _, name := range field.Names {
				parts = append(parts, name.Name+" "+typeStr)
			}
		}
	}
	return strings.Join(parts, ", ")
}

func exprToString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, expr)
	return buf.String()
}

func collectImportsFromExpr(expr ast.Expr, fileImports map[string]string, importSet map[string]bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if path, exists := fileImports[ident.Name]; exists {
			importSet[path] = true
		}
		return true
	})
}

func renderServiceFile(info *servicePackageInfo) (string, error) {
	t, err := template.New("service").Parse(serviceTpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, info); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type logicEntryData struct {
	Packages []string
}

func generateLogicEntry(srcPath string, pkgPaths []string) error {
	data := logicEntryData{Packages: pkgPaths}
	t, err := template.New("logic").Parse(logicTpl)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	outFile := filepath.Join(srcPath, "logic.go")
	if err := writeFormattedService(outFile, buf.String()); err != nil {
		return err
	}
	fmt.Printf("  generated: %s (logic entry)\n", outFile)
	return nil
}

func writeFormattedService(path, code string) error {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return os.WriteFile(path, []byte(code), 0644)
	}
	return os.WriteFile(path, formatted, 0644)
}
