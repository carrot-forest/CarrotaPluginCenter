package router

import (
	"carrota-plugin-center/controllers"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func Load(e *echo.Echo) {
	routes(e)
}

func routes(e *echo.Echo) {
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	apiVersionUrl := "/api/v1"

	e.GET(apiVersionUrl+"", controllers.IndexGET)
	e.GET(apiVersionUrl+"/", controllers.IndexGET)

	e.GET(apiVersionUrl+"/health", controllers.HealthGET)

	pluginGroup := e.Group(apiVersionUrl + "/plugin")
	{
		pluginGroup.POST("/register", controllers.PluginRegisterPOST)
		pluginGroup.GET("/list", controllers.PluginListGET)
	}

	messageGroup := e.Group(apiVersionUrl + "/message")
	{
		messageGroup.POST("", controllers.MessagePOST)
		messageGroup.POST("/", controllers.MessagePOST)
		messageGroup.POST("/send", controllers.MessageSendPOST)
	}
}
