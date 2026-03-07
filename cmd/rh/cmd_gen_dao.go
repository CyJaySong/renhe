package main

import (
	"fmt"
	"os"

	"github.com/cyjaysong/renhe/cmd/rh/internal/config"
	"github.com/cyjaysong/renhe/cmd/rh/internal/gen"
	"github.com/spf13/cobra"
)

var daoLink string

var genDaoCmd = &cobra.Command{
	Use:   "dao",
	Short: "Generate DAO, entity code from database tables (reads ./hack/config.yaml)",
	Long: `Generate DAO and entity code from database tables.

Reads configuration from ./hack/config.yaml in the project root.
Auto-detects Go module name from go.mod.

Configuration (hack/config.yaml):
  rh:
    gen:
      dao:
        - link:         PostgreSQL DSN (required)
                        e.g. "postgres://user:pass@host:port/dbname?sslmode=disable"
          schema:       Database schema (default: "public")
          tables:       Comma-separated table names or glob patterns to include (empty = all)
                        Supports: * (any chars), ? (single char). e.g. "user_*,config"
          tablesEx:     Comma-separated table names or glob patterns to exclude
                        Supports: * (any chars), ? (single char). e.g. "tmp_*,log_???"
          path:         Base output path (default: "./internal")
          tablePath:    Table metadata sub-path relative to path (default: "model/table")
          doPath:       DO sub-path relative to path (default: "model/do")
          entityPath:   Entity sub-path relative to path (default: "model/ent")
          jsonCase:     JSON tag naming: Snake / CamelLower / Camel (default: "Snake")
          typeMapping:  Map DB column types to custom Go types globally.
                        Each key is a DB type name (e.g. numeric, jsonb).
                        Value has "type" (Go type) and optional "import" (package path).
                        Example:
                          numeric:
                            type: "decimal.Decimal"
                            import: "github.com/shopspring/decimal"
          fieldMapping: Map specific table.column to custom Go types (higher priority than typeMapping).
                        Each key is "table_name.column_name".
                        Value has "type" (Go type) and optional "import" (package path).
                        Example:
                          user.other:
                            type: "map[string]any"

Multiple database configurations are supported as a YAML list.`,
	RunE: runGenDao,
}

func init() {
	genDaoCmd.Flags().StringVarP(&daoLink, "link", "l", "", "Override database link (optional, reads from hack/config.yaml by default)")
}

func runGenDao(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()

	module, err := config.DetectModule(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect go module: %w (ensure you run this command in the project root)", err)
	}

	cfg, err := config.Load(cwd + "/hack/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	daoConfigs := cfg.RH.Gen.Dao
	if len(daoConfigs) == 0 {
		return fmt.Errorf("no rh.gen.dao configuration found in hack/config.yaml")
	}

	for i, dc := range daoConfigs {
		link := dc.Link
		if i == 0 && daoLink != "" {
			link = daoLink
		}

		dsn, err := config.ParseLink(link)
		if err != nil {
			return fmt.Errorf("config[%d]: %w", i, err)
		}

		genCfg := gen.DaoConfig{
			DSN:        dsn,
			Schema:     dc.Schema,
			Tables:     dc.Tables,
			TablesEx:   dc.TablesEx,
			Path:       dc.Path,
			TablePath:  dc.TablePath,
			DoPath:     dc.DoPath,
			EntityPath: dc.EntityPath,
			JsonCase:   dc.JsonCase,
			Module:     module,
		}
		// 转换 config.TypeMappingItem → gen.TypeMappingItem
		if len(dc.TypeMapping) > 0 {
			genCfg.TypeMapping = make(map[string]gen.TypeMappingItem, len(dc.TypeMapping))
			for k, v := range dc.TypeMapping {
				genCfg.TypeMapping[k] = gen.TypeMappingItem{Type: v.Type, Import: v.Import}
			}
		}
		if len(dc.FieldMapping) > 0 {
			genCfg.FieldMapping = make(map[string]gen.TypeMappingItem, len(dc.FieldMapping))
			for k, v := range dc.FieldMapping {
				genCfg.FieldMapping[k] = gen.TypeMappingItem{Type: v.Type, Import: v.Import}
			}
		}

		fmt.Printf("=== Processing config[%d] ===\n", i)
		if err := gen.RunDao(genCfg); err != nil {
			return fmt.Errorf("config[%d]: %w", i, err)
		}
	}

	return nil
}
