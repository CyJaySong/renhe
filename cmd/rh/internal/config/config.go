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

type GenDaoConfig struct {
	Link       string `yaml:"link"`
	Tables     string `yaml:"tables"`
	TablesEx   string `yaml:"tablesEx"`
	Schema     string `yaml:"schema"`
	Path       string `yaml:"path"`
	TablePath  string `yaml:"tablePath"`
	DoPath     string `yaml:"doPath"`
	EntityPath string `yaml:"entityPath"`
	JsonCase   string `yaml:"jsonCase"`
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
