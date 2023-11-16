package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
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

// Plugin Center 请求结构体
type MessageInfo struct {
	MessageID string      `json:"message_id"`
	Agent     string      `json:"agent"`
	GroupID   string      `json:"group_id"`
	GroupName string      `json:"group_name"`
	UserID    string      `json:"user_id"`
	UserName  string      `json:"user_name"`
	Time      int64       `json:"time"`
	Message   string      `json:"message"`
	IsMention bool        `json:"is_mention"`
	Param     interface{} `json:"param"`
}

type MessageResponse struct {
	IsReply bool     `json:"is_reply"`
	Message []string `json:"message"`
}

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
	if !strings.Contains(message.Message, "占卜") {
		return c.JSON(http.StatusOK, MessageResponse{
			IsReply: false,
			Message: []string{},
		})
	}

	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(100) + 1
	const (
		CarrotDiceSuccessFullPoint = "！！100% 概率事件！！你就是命运之神本神！！卡洛要好好记录下来..."
		CarrotDiceSuccessGold      = "占卜的结果是非常大概率！信心满满的往前走吧！"
		CarrotDiceSuccessSilver    = "是很难避免的事情哦~愿命运之神与你同在"
		CarrotDiceSuccessBronze    = "卡洛认为基本可以放轻松啦"
		CarrotDiceFailedGold       = "卡洛只能看见命运的天平在不停的摇摆摇摆摇摆~"
		CarrotDiceFailedSilver     = "水晶球里徘徊着一大团迷雾...情况看起来不太对劲，卡洛奉劝你最好小心"
		CarrotDiceFailedZeroPoint  = "哼哼，卡洛的占卜术显示这是不可能发生的事哦！"
	)
	result := "卡洛" + message.Message + "的结果是："
	if randomInt == 100 {
		result += CarrotDiceSuccessFullPoint
	} else if randomInt >= 85 {
		result += CarrotDiceSuccessGold
	} else if randomInt >= 70 {
		result += CarrotDiceSuccessSilver
	} else if randomInt >= 55 {
		result += CarrotDiceSuccessBronze
	} else if randomInt >= 40 {
		result += CarrotDiceFailedGold
	} else if randomInt >= 25 {
		result += CarrotDiceFailedSilver
	} else {
		result += CarrotDiceFailedZeroPoint
	}

	return c.JSON(http.StatusOK, MessageResponse{
		IsReply: true,
		Message: []string{result},
	})
}

type PluginParam struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
type PluginInfo struct {
	ID          string        `json:"id"          `
	Name        string        `json:"name"        `
	Author      string        `json:"author"      `
	Description string        `json:"description" `
	Prompt      string        `json:"prompt"      `
	Params      []PluginParam `json:"param"       `
	Format      []string      `json:"format"      `
	Example     []string      `json:"example"     `
	Url         string        `json:"url"         `
}

func Register(pluginCenterEndpoint string, pluginEndpoint string) error {
	/****************** 注册插件 ******************/
	jsonStr, _ := json.Marshal(PluginInfo{
		ID:          "divine",
		Name:        "占卜",
		Author:      "ligen131",
		Description: "如果语句中带有占卜，请触发这个插件。这个接口不需要传入任何参数，将返回一个随机数",
		Prompt:      "占卜",
		Params:      []PluginParam{},
		Format: []string{
			"占卜我明天能不能吃饱饭",
		},
		Example: []string{
			"不调用",
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
