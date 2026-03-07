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
