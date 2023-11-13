package config

import (
	"carrota-plugin-center/model"
	"carrota-plugin-center/utils/logs"

	"carrota-plugin-center/controllers/auth"
	"carrota-plugin-center/shared/server"
	"carrota-plugin-center/shared/service"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"go.uber.org/zap"
)

func ConfigLoadInit() {
	// 设置选项支持 ENV 解析
	config.WithOptions(config.ParseEnv)

	// 添加驱动程序以支持 yaml 内容解析
	config.AddDriver(yaml.Driver)
	config.WithOptions(func(opt *config.Options) {
		opt.DecoderConfig.TagName = "config"
	})
}

type YamlConfiguration struct {
	Server         server.Server                `config:"server"`
	Database       model.Database               `config:"database"`
	Authorization  auth.Authorization           `config:"Authorization"`
	CarrotaService service.CarrotaServiceConfig `config:"carrota-service"`
}

func YamlConfigLoad(path string) (YamlConfiguration, error) {
	configuration := YamlConfiguration{}
	err := config.LoadFiles(path)
	if err != nil {
		logs.Error("Read config file from "+path+"failed. ", zap.Error(err))
		return configuration, err
	}

	err = config.Decode(&configuration)
	if err != nil {
		logs.Error("Decode config file from "+path+"failed. ", zap.Error(err))
		return configuration, err
	}
	logs.Info("Read config file from "+path, zap.Any("configuration", configuration))

	return configuration, nil
}
