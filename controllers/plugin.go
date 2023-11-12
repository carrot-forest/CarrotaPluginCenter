package controllers

import (
	"carrota-plugin-center/model"
	"carrota-plugin-center/utils/logs"

	"github.com/labstack/echo/v4"
)

func PluginRegisterPOST(c echo.Context) error {
	logs.Debug("POST /plugin/register")

	plugin := model.PluginInfo{}
	_ok, err := Bind(c, &plugin)
	if !_ok {
		return err
	}

	err = model.CreatePluginRegisterRecord(plugin)
	if err != nil {
		return ResponseInternalServerError(c, "Create PluginRegisterRecord failed.", err)
	}
	return ResponseOK(c, "ok")
}

func PluginListGET(c echo.Context) error {
	logs.Debug("GET /plugin/list")

	plugins, err := model.FindPluginList()
	if err != nil {
		return ResponseInternalServerError(c, "Find plugin list failed.", err)
	}
	return ResponseOK(c, plugins)
}