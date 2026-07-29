package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
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
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	info := &servicePackageInfo{
		PkgName:    pkgName,
		PascalName: ToPascalCase(pkgName),
	}

	importSet := make(map[string]bool)

	for _, pkg := range pkgs {
		fileNames := make([]string, 0, len(pkg.Files))
		for name := range pkg.Files {
			fileNames = append(fileNames, name)
		}
		sort.Strings(fileNames)
		for _, name := range fileNames {
			file := pkg.Files[name]
			for _, imp := range collectFileImports(file) {
				if !importSet[imp] {
					importSet[imp] = true
					info.Imports = append(info.Imports, imp)
				}
			}
			for _, method := range extractMethods(fset, file) {
				info.Methods = append(info.Methods, method)
			}
		}
	}

	return info, nil
}

func collectFileImports(file *ast.File) []string {
	var result []string
	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "_" {
			continue
		}
		var decl string
		if imp.Name != nil {
			decl = imp.Name.Name + " " + imp.Path.Value
		} else {
			decl = imp.Path.Value
		}
		result = append(result, decl)
	}
	return result
}

func extractMethods(fset *token.FileSet, file *ast.File) []serviceMethod {
	var methods []serviceMethod

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || !fn.Name.IsExported() {
			continue
		}

		params := formatFieldList(fset, fn.Type.Params)
		results := formatFieldList(fset, fn.Type.Results)

		resultStr := results
		if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
			needParen := len(fn.Type.Results.List) > 1
			if !needParen {
				for _, field := range fn.Type.Results.List {
					if len(field.Names) > 0 {
						needParen = true
						break
					}
				}
			}
			if needParen {
				resultStr = "(" + results + ")"
			}
		}

		methods = append(methods, serviceMethod{
			Name:    fn.Name.Name,
			Params:  params,
			Results: resultStr,
		})
	}
	return methods
}

func formatFieldList(fset *token.FileSet, fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}

	var parts []string
	for _, field := range fl.List {
		typeStr := exprToString(fset, field.Type)
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
	formatted, err := imports.Process(path, []byte(code), nil)
	if err != nil {
		return fmt.Errorf("format %s: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0644)
}
