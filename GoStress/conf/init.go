package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	confMap = map[string]interface{}{} // 存储所有配置文件信息
)

// 本地静态配置文件加载
func staicConfInit(path string) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var staticViper *viper.Viper
		if !info.IsDir() && filepath.Ext(path) == ".toml" {
			staticViper = viper.New()
			staticViper.SetConfigFile(path)

			err := staticViper.ReadInConfig()
			if err != nil {
				return err
			}

			conf := staticViper.AllSettings()
			val, ok := conf["service_name"]
			if !ok {
				return fmt.Errorf("wrong file conf format: %s", path)
			}

			confMap[val.(string)] = conf

		}

		return nil
	})

	if err != nil {
		panic(err)
	}

}

func Init(path string) {
	staicConfInit(path)
}
