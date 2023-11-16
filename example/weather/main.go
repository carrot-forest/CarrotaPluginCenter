package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var logs *zap.Logger

const path = "config.yml" // 配置文件路径，默认是当前文件夹下的 config.yml

type ServerConfig struct {
	Hostname string `config:"hostname"`
	Port     int    `config:"port"`
}
type DatabaseConfig struct {
	Hostname string `config:"hostname"`
	Port     int    `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	DbName   string `config:"dbName"`
	SslMode  bool   `config:"sslMode"`
	TimeZone string `config:"timeZone"`
}
type YamlConfiguration struct {
	Server               ServerConfig   `config:"server"`
	Database             DatabaseConfig `config:"database"`
	PluginCenterEndpoint string         `config:"plugin-center-endpoint"`
	PluginEndpoint       string         `config:"plugin-endpoint"`
	QWeatherToken        string         `config:"qweather-token"`
}

// 读取配置文件
func YamlConfigLoad(path string) (YamlConfiguration, error) {
	// 设置选项支持 ENV 解析
	config.WithOptions(config.ParseEnv)

	// 添加驱动程序以支持 yaml 内容解析
	config.AddDriver(yaml.Driver)
	config.WithOptions(func(opt *config.Options) {
		// 设置 struct tag 名称
		opt.DecoderConfig.TagName = "config"
	})

	// 文件读取
	configuration := YamlConfiguration{}
	err := config.LoadFiles(path)
	if err != nil {
		logs.Error("Read config file from "+path+"failed. ", zap.Error(err))
		return configuration, err
	}

	// 解析 YAML
	err = config.Decode(&configuration)
	if err != nil {
		logs.Error("Decode config file from "+path+"failed. ", zap.Error(err))
		return configuration, err
	}
	logs.Info("Read config file from "+path, zap.Any("configuration", configuration))

	return configuration, nil
}

// 将请求绑定到结构体
func Bind(c echo.Context, obj interface{}) (bool, error) {
	err := c.Bind(&obj)
	if err != nil {
		logs.Warn("Failed to parse request data.", zap.Error(err))
		return false, c.JSON(http.StatusBadRequest, "Failed to parse request data.")
	}
	logs.Debug("Parsed struct:", zap.Any("obj", obj))
	return true, nil
}

type Weather struct {
	City     string `json:"city"`
	DayDelta int    `json:"NumberOfDaysFromToday"`
}

type CityLookupLocation struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Adm2 string `json:"adm2"`
	Adm1 string `json:"adm1"`
}

type CityLookupResponse struct {
	Location []CityLookupLocation `json:"location"`
}

type WeatherDaily struct {
	FxDate       string `json:"fxDate"`
	TempMax      string `json:"tempMax"`
	TempMin      string `json:"tempMin"`
	TextDay      string `json:"textDay"`
	TextNight    string `json:"textNight"`
	WindDirDay   string `json:"windDirDay"`
	WindDirNight string `json:"windDirNight"`
}

type WeatherPredictResponse struct {
	Daily []WeatherDaily `json:"daily"`
}

var qweatherToken string

// Plugin Center 请求结构体
type MessageInfo struct {
	MessageID string  `json:"message_id"`
	Agent     string  `json:"agent"`
	GroupID   string  `json:"group_id"`
	GroupName string  `json:"group_name"`
	UserID    string  `json:"user_id"`
	UserName  string  `json:"user_name"`
	Time      int64   `json:"time"`
	Message   string  `json:"message"`
	IsMention bool    `json:"is_mention"`
	Param     Weather `json:"param"`
}

type MessageResponse struct {
	IsReply bool     `json:"is_reply"`
	Message []string `json:"message"`
}

var messageRecord = make(map[string][]MessageInfo)

const repeatCount = 3

func process(c echo.Context) error {
	logs.Debug("POST /")

	// 将请求绑定到结构体
	message := MessageInfo{}
	_ok, err := Bind(c, &message)
	if !_ok {
		// 失败返回 400 Bad Request
		return err
	}

	/*************** 插件处理逻辑 ***************/
	city := message.Param.City
	dayDelta := message.Param.DayDelta
	if city == "" {
		city = "武汉"
	}
	if dayDelta < 0 {
		dayDelta = 0
	}
	if dayDelta >= 7 {
		dayDelta = 7
	}

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("https://geoapi.qweather.com/v2/city/lookup?key=%s&location=%s", qweatherToken, url.QueryEscape(city)),
		nil)
	logs.Debug("GET QWeather City Lookup Endpoint", zap.String("url", req.URL.String()))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp == nil {
			logs.Error("GET QWeather City Lookup Endpoint failed", zap.Error(err))
		} else {
			logs.Error("GET QWeather City Lookup Endpoint failed", zap.Int("StatusCode", resp.StatusCode), zap.Error(err))
		}
		return err
	}

	cityLookupResponse := CityLookupResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cityLookupResponse)
	if err != nil {
		logs.Error("Decode cityLookupResponse failed", zap.Error(err))
		return err
	}
	resp.Body.Close()
	logs.Debug("cityLookupResponse", zap.Any("cityLookupResponse", cityLookupResponse))

	cityID := ""
	for _, location := range cityLookupResponse.Location {
		if location.Adm2 == city {
			cityID = location.ID
		}
	}
	if cityID == "" {
		cityID = cityLookupResponse.Location[0].ID
	}

	req, _ = http.NewRequest("GET",
		fmt.Sprintf("https://devapi.qweather.com/v7/weather/7d?key=%s&location=%s", qweatherToken, cityID),
		nil)
	logs.Debug("GET QWeather Endpoint", zap.String("url", req.URL.String()))
	req.Header.Set("Content-Type", "application/json")
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp == nil {
			logs.Error("GET QWeather Endpoint failed", zap.Error(err))
		} else {
			logs.Error("GET QWeather Endpoint failed", zap.Int("StatusCode", resp.StatusCode), zap.Error(err))
		}
		return err
	}

	weatherPredictResponse := WeatherPredictResponse{}
	err = json.NewDecoder(resp.Body).Decode(&weatherPredictResponse)
	if err != nil {
		logs.Error("Decode weatherPredictResponse failed", zap.Error(err))
		return err
	}
	resp.Body.Close()

	weatherReply := weatherPredictResponse.Daily[dayDelta]
	return c.JSON(http.StatusOK, MessageResponse{
		IsReply: true,
		Message: []string{
			fmt.Sprintf("%s的天气情况如下：气温%s℃~%s℃，白天%s，风向%s，夜晚%s，风向%s",
				weatherReply.FxDate, weatherReply.TempMin, weatherReply.TempMax,
				weatherReply.TextDay, weatherReply.WindDirDay,
				weatherReply.TextNight, weatherReply.WindDirNight),
		},
	})
}

type PluginParam struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
type PluginInfo struct {
	ID          string        `json:"id"         `
	Name        string        `json:"name"       `
	Author      string        `json:"author"     `
	Description string        `json:"description"`
	Prompt      string        `json:"prompt"     `
	Params      []PluginParam `json:"param"      `
	Format      []string      `json:"format"     `
	Example     []string      `json:"example"    `
	Url         string        `json:"url"        `
}

func Register(pluginCenterEndpoint string, pluginEndpoint string) error {
	/****************** 注册插件 ******************/
	jsonStr, _ := json.Marshal(PluginInfo{
		ID:          "weather",
		Name:        "天气预报",
		Author:      "ligen131",
		Description: "天气查询：传入一个参数即地点，将返回这个地点的天气。如果你认为需要进行天气查询，但语句中没有给出地点，请传入“武汉”。",
		Prompt:      "复读机",
		Params: []PluginParam{
			{
				Key:         "city",
				Type:        "string",
				Description: "城市名",
			},
			{
				Key:         "NumberOfDaysFromToday",
				Type:        "int",
				Description: "距今天的天数，如今天为0，明天为1，三天后为3",
			},
		},
		Format: []string{
			"${city}的天气怎么样",
			"${NumberOfDaysFromToday}的天气怎么样",
			"${NumberOfDaysFromToday}${city}的天气怎么样",
		},
		Example: []string{
			"调用 武汉 0",
		},
		Url: pluginEndpoint,
	})
	req, _ := http.NewRequest("POST", pluginCenterEndpoint+"/plugin/register", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp == nil {
			logs.Error("POST Plugin Center Register Endpoint failed", zap.Error(err))
		} else {
			logs.Error("POST Plugin Center Register Endpoint failed", zap.Int("StatusCode", resp.StatusCode), zap.Error(err))
		}
		return err
	}
	return nil
}

func main() {
	// logger 初始化
	logs, _ = zap.NewDevelopment()

	// 读取配置文件
	config, err := YamlConfigLoad(path)
	if err != nil {
		panic(err)
	}

	qweatherToken = config.QWeatherToken

	err = Register(config.PluginCenterEndpoint, config.PluginEndpoint)
	if err != nil {
		logs.Error("Register failed", zap.Error(err))
	}
	// 每隔 1 分钟自动注册一次插件
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			err := Register(config.PluginCenterEndpoint, config.PluginEndpoint)
			if err != nil {
				logs.Error("Register failed", zap.Error(err))
			}
		}
	}()

	// 配置后端服务
	e := echo.New()
	e.POST("/", process)

	// 启动！
	e.Logger.Fatal(e.Start(config.Server.Hostname + ":" + fmt.Sprint(config.Server.Port)))
}
