package gen

import (
	"database/sql"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TypeMappingItem 类型映射配置项。
type TypeMappingItem struct {
	Type   string // Go 类型名，如 decimal.Decimal
	Import string // 需要导入的包路径
	PkgAs  string // 导入包别名，如 decimalx
}

type DaoConfig struct {
	DSN           string
	Schema        string
	Tables        string
	TablesEx      string
	Path          string
	TablePath     string
	DoPath        string
	EntityPath    string
	JsonCase      string
	Module        string
	EntityFieldEx string                     // 生成 entity 时排除的字段
	TypeMapping   map[string]TypeMappingItem // 按数据库类型名全局映射
	FieldMapping  map[string]TypeMappingItem // 按 表名.字段名 精确映射
}

type columnInfo struct {
	Name             string
	DataType         string
	UdtName          string
	IsNullable       string
	IsPrimary        bool
	AutoIncrement    bool // 是否自增列（identity 或序列默认值）
	HasDefault       bool
	DefaultValue     string // 原始默认值表达式
	Comment          string
	MaxLength        int    // character_maximum_length
	NumericPrecision int    // numeric_precision
	NumericScale     int    // numeric_scale
	ElementType      string // 数组元素的 udt_name（仅 ARRAY 类型）
}

func RunDao(cfg DaoConfig) error {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	tables, err := queryTables(db, cfg.Schema)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	tables = filterTables(tables, cfg.Tables, cfg.TablesEx)

	if len(tables) == 0 {
		fmt.Println("No tables found to generate.")
		return nil
	}

	tableDir := filepath.Join(cfg.Path, cfg.TablePath)
	tableInternalDir := filepath.Join(tableDir, "internal")
	doDir := filepath.Join(cfg.Path, cfg.DoPath)
	entityDir := filepath.Join(cfg.Path, cfg.EntityPath)

	for _, dir := range []string{tableDir, tableInternalDir, doDir, entityDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	tableImportPath := cfg.Module + "/" + filepath.ToSlash(filepath.Join(cfg.Path, cfg.TablePath))
	entityImportPath := cfg.Module + "/" + filepath.ToSlash(filepath.Join(cfg.Path, cfg.EntityPath))

	for _, table := range tables {
		cols, err := queryColumns(db, cfg.Schema, table)
		if err != nil {
			return fmt.Errorf("failed to get columns for %s: %w", table, err)
		}

		// Generate do
		doCode := generateDoCode(table, cols, filepath.Base(cfg.DoPath))
		doFile := filepath.Join(doDir, table+".go")
		if err := writeFormatted(doFile, doCode); err != nil {
			return fmt.Errorf("failed to write %s: %w", doFile, err)
		}
		fmt.Printf("  generated: %s (do)\n", doFile)

		// Generate entity
		entityCode := generateEntityCode(table, cols, filepath.Base(cfg.EntityPath), cfg.JsonCase, cfg.TypeMapping, cfg.FieldMapping, cfg.EntityFieldEx)
		entityFile := filepath.Join(entityDir, table+".go")
		if err := writeFormatted(entityFile, entityCode); err != nil {
			return fmt.Errorf("failed to write %s: %w", entityFile, err)
		}
		fmt.Printf("  generated: %s (entity)\n", entityFile)

		// Generate table/internal
		internalCode := generateDaoInternal(table, cols)
		internalFile := filepath.Join(tableInternalDir, table+".go")
		if err := writeFormatted(internalFile, internalCode); err != nil {
			return fmt.Errorf("failed to write %s: %w", internalFile, err)
		}
		fmt.Printf("  generated: %s (table/internal)\n", internalFile)

		// Generate table
		tableCode := generateTableFile(table, filepath.Base(cfg.TablePath), tableImportPath, entityImportPath)
		tableFile := filepath.Join(tableDir, table+".go")
		if err := writeFormatted(tableFile, tableCode); err != nil {
			return fmt.Errorf("failed to write %s: %w", tableFile, err)
		}
		fmt.Printf("  generated: %s (table)\n", tableFile)
	}

	fmt.Printf("\nGeneration completed! %d table(s) processed.\n", len(tables))
	return nil
}

func writeFormatted(path, code string) error {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return os.WriteFile(path, []byte(code), 0644)
	}
	return os.WriteFile(path, formatted, 0644)
}

func filterTables(all []string, include, exclude string) []string {
	allSet := make(map[string]bool, len(all))
	for _, t := range all {
		allSet[t] = true
	}

	result := all
	if include != "" {
		patterns := splitPatterns(include)
		matched := make(map[string]bool)
		for _, p := range patterns {
			if isGlobPattern(p) {
				for _, t := range all {
					if ok, _ := path.Match(p, t); ok {
						matched[t] = true
					}
				}
			} else {
				if allSet[p] {
					matched[p] = true
				} else {
					fmt.Printf("  warning: table %q does not exist, skipped\n", p)
				}
			}
		}
		var filtered []string
		for _, t := range all {
			if matched[t] {
				filtered = append(filtered, t)
			}
		}
		result = filtered
	}

	if exclude != "" {
		patterns := splitPatterns(exclude)
		excluded := make(map[string]bool)
		for _, p := range patterns {
			if isGlobPattern(p) {
				for _, t := range result {
					if ok, _ := path.Match(p, t); ok {
						excluded[t] = true
					}
				}
			} else {
				excluded[p] = true
			}
		}
		var filtered []string
		for _, t := range result {
			if !excluded[t] {
				filtered = append(filtered, t)
			}
		}
		result = filtered
	}
	return result
}

func splitPatterns(s string) []string {
	parts := strings.Split(s, ",")
	patterns := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			patterns = append(patterns, p)
		}
	}
	return patterns
}

func isGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func queryTables(db *sql.DB, schema string) ([]string, error) {
	rows, err := db.Query(
		"SELECT table_name FROM information_schema.tables "+
			"WHERE table_schema = $1 AND table_type = 'BASE TABLE' ORDER BY table_name",
		schema,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func queryColumns(db *sql.DB, schema, table string) ([]columnInfo, error) {
	query := `
SELECT
	c.column_name,
	c.data_type,
	c.udt_name,
	c.is_nullable,
	COALESCE(c.column_default, '') != '' AS has_default,
	COALESCE(c.column_default, '') AS default_value,
	-- 自增判定：identity 列（GENERATED ... AS IDENTITY）或 serial 序列默认值
	(c.is_identity = 'YES' OR COALESCE(c.column_default, '') LIKE 'nextval(%') AS is_auto_increment,
	COALESCE(
		(SELECT true FROM information_schema.table_constraints tc
		 JOIN information_schema.key_column_usage kcu
		   ON tc.constraint_name = kcu.constraint_name
		  AND tc.table_schema = kcu.table_schema
		 WHERE tc.constraint_type = 'PRIMARY KEY'
		   AND tc.table_schema = $1
		   AND tc.table_name = $2
		   AND kcu.column_name = c.column_name
		 LIMIT 1),
		false
	) AS is_primary,
	COALESCE(pgd.description, '') AS comment,
	COALESCE(c.character_maximum_length, 0) AS max_length,
	COALESCE(c.numeric_precision, 0) AS numeric_precision,
	COALESCE(c.numeric_scale, 0) AS numeric_scale,
	COALESCE(e.udt_name, '') AS element_type
FROM information_schema.columns c
LEFT JOIN pg_catalog.pg_statio_all_tables st
	ON st.schemaname = c.table_schema AND st.relname = c.table_name
LEFT JOIN pg_catalog.pg_description pgd
	ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
LEFT JOIN information_schema.element_types e
	ON e.object_catalog = c.table_catalog
	AND e.object_schema = c.table_schema
	AND e.object_name = c.table_name
	AND e.object_type = 'TABLE'
	AND e.collection_type_identifier = c.dtd_identifier
WHERE c.table_schema = $1 AND c.table_name = $2
ORDER BY c.ordinal_position`

	rows, err := db.Query(query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []columnInfo
	for rows.Next() {
		var c columnInfo
		if err := rows.Scan(
			&c.Name, &c.DataType, &c.UdtName, &c.IsNullable,
			&c.HasDefault, &c.DefaultValue, &c.AutoIncrement, &c.IsPrimary, &c.Comment,
			&c.MaxLength, &c.NumericPrecision, &c.NumericScale, &c.ElementType,
		); err != nil {
			return nil, err
		}
		columns = append(columns, c)
	}
	return columns, nil
}

func pgTypeToGo(c columnInfo) string {
	nullable := c.IsNullable == "YES"
	var goType string
	switch strings.ToLower(c.DataType) {
	case "smallint":
		goType = "int16"
	case "integer":
		goType = "int32"
	case "bigint":
		goType = "int64"
	case "real":
		goType = "float32"
	case "double precision":
		goType = "float64"
	case "numeric":
		goType = "string"
	case "character varying", "character", "text":
		goType = "string"
	case "boolean":
		goType = "bool"
	case "timestamp without time zone", "timestamp with time zone", "date":
		goType = "time.Time"
	case "time without time zone", "time with time zone":
		goType = "string"
	case "bytea":
		goType = "[]byte"
	case "json", "jsonb":
		goType = "json.RawMessage"
	case "uuid":
		goType = "string"
	case "inet", "cidr", "macaddr":
		goType = "string"
	case "interval":
		goType = "string"
	case "array", "ARRAY":
		goType = pgArrayElementGoType(c.ElementType)
	case "user-defined":
		goType = "string"
	default:
		goType = "string"
	}
	if nullable && goType != "string" && goType != "[]byte" && goType != "json.RawMessage" && !strings.HasPrefix(goType, "[]") {
		return "*" + goType
	}
	return goType
}

// pgArrayElementGoType 根据数组元素的 udt_name 返回 Go 切片类型。
func pgArrayElementGoType(elemUdt string) string {
	switch strings.ToLower(elemUdt) {
	case "int2":
		return "[]int16"
	case "int4":
		return "[]int32"
	case "int8":
		return "[]int64"
	case "float4":
		return "[]float32"
	case "float8", "numeric":
		return "[]float64"
	case "bool":
		return "[]bool"
	case "varchar", "text", "bpchar", "uuid":
		return "[]string"
	default:
		return "[]string"
	}
}

func toJsonCase(name, jsonCase string) string {
	switch jsonCase {
	case "CamelLower":
		pascal := ToPascalCase(name)
		if len(pascal) == 0 {
			return name
		}
		runes := []rune(pascal)
		runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
		return string(runes)
	case "Camel":
		return ToPascalCase(name)
	case "Snake":
		return name
	default:
		return name
	}
}
