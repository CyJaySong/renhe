package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RH RHConfig `yaml:"rh"`
}

type RHConfig struct {
	Gen GenConfig `yaml:"gen"`
}

type GenConfig struct {
	Dao     []GenDaoConfig   `yaml:"dao"`
	Service GenServiceConfig `yaml:"service"`
}

type GenServiceConfig struct {
	SrcPath string `yaml:"srcPath"`
	DstPath string `yaml:"dstPath"`
}

// TypeMappingItem 类型映射项，配置数据库字段类型或具体字段对应的 Go 类型。
type TypeMappingItem struct {
	Type   string `yaml:"type"`   // Go 类型名，如 decimal.Decimal
	Import string `yaml:"import"` // 需要导入的包路径，如 github.com/shopspring/decimal
	PkgAs  string `yaml:"pkgAs"`  // 导入包别名，如 decimalx
}

type GenDaoConfig struct {
	Link          string                     `yaml:"link"`
	Tables        string                     `yaml:"tables"`
	TablesEx      string                     `yaml:"tablesEx"`
	Schema        string                     `yaml:"schema"`
	Path          string                     `yaml:"path"`
	TablePath     string                     `yaml:"tablePath"`
	DoPath        string                     `yaml:"doPath"`
	EntityPath    string                     `yaml:"entityPath"`
	JsonCase      string                     `yaml:"jsonCase"`
	EntityFieldEx string                     `yaml:"entityFieldEx"` // 生成 entity 时排除的字段，格式: "表名.字段名" 或 "*.字段名"，逗号分隔
	TypeMapping   map[string]TypeMappingItem `yaml:"typeMapping"`   // 按数据库类型名全局映射，如 numeric → decimal.Decimal
	FieldMapping  map[string]TypeMappingItem `yaml:"fieldMapping"`  // 按 表名.字段名 精确映射，优先级高于 typeMapping
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	for i := range cfg.RH.Gen.Dao {
		applyDaoDefaults(&cfg.RH.Gen.Dao[i])
	}
	applyServiceDefaults(&cfg.RH.Gen.Service)
	return &cfg, nil
}

func applyDaoDefaults(c *GenDaoConfig) {
	if c.Schema == "" {
		c.Schema = "public"
	}
	if c.Path == "" {
		c.Path = "./internal"
	}
	if c.TablePath == "" {
		c.TablePath = "model/table"
	}
	if c.DoPath == "" {
		c.DoPath = "model/do"
	}
	if c.EntityPath == "" {
		c.EntityPath = "model/ent"
	}
	if c.JsonCase == "" {
		c.JsonCase = "Snake"
	}
}

func applyServiceDefaults(c *GenServiceConfig) {
	if c.SrcPath == "" {
		c.SrcPath = "internal/logic"
	}
	if c.DstPath == "" {
		c.DstPath = "internal/service"
	}
}

// DetectModule reads the go.mod file in the given directory and returns the module path.
func DetectModule(dir string) (string, error) {
	f, err := os.Open(dir + "/go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(line[7:]), nil
		}
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

// ParseLink validates and returns a standard PostgreSQL DSN.
// Supported format: "postgres://user:pass@host:port/dbname?sslmode=disable"
func ParseLink(link string) (string, error) {
	if strings.HasPrefix(link, "postgres://") || strings.HasPrefix(link, "postgresql://") {
		return link, nil
	}
	return "", fmt.Errorf("unsupported link format, expected standard DSN like: postgres://user:pass@host:port/dbname?sslmode=disable\n  got: %s", link)
}
