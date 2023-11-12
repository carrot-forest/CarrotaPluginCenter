package main

import (
	"carrota-plugin-center/controllers/auth"
	"carrota-plugin-center/model"
	"carrota-plugin-center/shared/config"
	"carrota-plugin-center/shared/server"
)

func main() {
	config.ConfigLoadInit()
	configuration, err := config.YamlConfigLoad("config.yml")
	if err != nil {
		panic(err)
	}

	err = model.Connect(configuration.Database)
	if err != nil {
		panic(err)
	}

	err = model.InitModel()
	if err != nil {
		panic(err)
	}

	err = auth.InitAuthorization(configuration.Authorization)
	if err != nil {
		panic(err)
	}

	err = server.Run(configuration.Server)
	if err != nil {
		panic(err)
	}
}
