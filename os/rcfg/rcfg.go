// Package rcfg 提供全局配置管理，基于 viper 实现 YAML 配置文件的读取与解析。
// 配置文件按以下优先级搜索：manifest/config > manifest > config > 当前目录。
package rcfg

import (
	"log"
	"strings"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// YamlTagOption 让 viper.Unmarshal 使用 yaml tag 而非默认的 mapstructure tag。
var YamlTagOption = viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
	dc.TagName = "yaml"
})

var (
	instance *viper.Viper
	once     sync.Once
)

// Cfg 返回全局 viper 实例（单例）。
// 可传入路径片段获取子配置，例如 Cfg("database", "default") 等价于 viper.Sub("database.default")。
// 首次调用时自动读取配置文件，读取失败将 Fatal。
func Cfg(paths ...string) *viper.Viper {
	once.Do(func() {
		v := viper.New()
		v.SetConfigType("yaml")
		v.SetConfigName("config")

		v.AddConfigPath("manifest/config")
		v.AddConfigPath("manifest")
		v.AddConfigPath("config")
		v.AddConfigPath(".")
		if err := v.ReadInConfig(); err != nil {
			log.Fatalf("配置文件读取出错：%s \n", err)
		}
		instance = v
	})
	switch {
	case len(paths) == 0:
		return instance
	default:
		return instance.Sub(strings.Join(paths, "."))
	}
}
