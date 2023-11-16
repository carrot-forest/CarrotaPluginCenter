package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"carrota-plugin-homework/logs"
	"carrota-plugin-homework/model"
)

const path = "config.yml" // 配置文件路径，默认是当前文件夹下的 config.yml

type ServerConfig struct {
	Hostname string `config:"hostname"`
	Port     int    `config:"port"`
}
type YamlConfiguration struct {
	Server               ServerConfig   `config:"server"`
	Database             model.Database `config:"database"`
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
		logs.Logs.Error("Read config file from "+path+"failed. ", zap.Error(err))
		return configuration, err
	}

	// 解析 YAML
	err = config.Decode(&configuration)
	if err != nil {
		logs.Logs.Error("Decode config file from "+path+"failed. ", zap.Error(err))
		return configuration, err
	}
	logs.Logs.Info("Read config file from "+path, zap.Any("configuration", configuration))

	return configuration, nil
}

// 将请求绑定到结构体
func Bind(c echo.Context, obj interface{}) (bool, error) {
	err := c.Bind(&obj)
	if err != nil {
		logs.Logs.Warn("Failed to parse request data.", zap.Error(err))
		return false, c.JSON(http.StatusBadRequest, "Failed to parse request data.")
	}
	logs.Logs.Debug("Parsed struct:", zap.Any("obj", obj))
	return true, nil
}

type HomeworkParam struct {
	Subject       string `json:"subject"`
	IsAddHomework string `json:"isAddHomework"`
	Content       string `json:"Content"`
	Deadline      string `json:"Deadline"`
}

// Plugin Center 请求结构体
type MessageInfo struct {
	MessageID string        `json:"message_id"`
	Agent     string        `json:"agent"`
	GroupID   string        `json:"group_id"`
	GroupName string        `json:"group_name"`
	UserID    string        `json:"user_id"`
	UserName  string        `json:"user_name"`
	Time      int64         `json:"time"`
	Message   string        `json:"message"`
	IsMention bool          `json:"is_mention"`
	Param     HomeworkParam `json:"param"`
}

type MessageResponse struct {
	IsReply bool     `json:"is_reply"`
	Message []string `json:"message"`
}

func process(c echo.Context) error {
	logs.Logs.Debug("POST /")

	// 将请求绑定到结构体
	message := MessageInfo{}
	_ok, err := Bind(c, &message)
	if !_ok {
		// 失败返回 400 Bad Request
		return err
	}

	/*************** 插件处理逻辑 ***************/
	content := message.Param.Content + "，截止时间：" + message.Param.Deadline
	if message.Param.IsAddHomework == "True" || message.Param.IsAddHomework == "true" {
		err := model.CreateHomeworkRecord(model.Homework{
			Subject: message.Param.Subject,
			Content: content,
		})
		if err != nil {
			logs.Logs.Error("Create homework record failed.", zap.Error(err))
			return c.JSON(http.StatusOK, MessageResponse{
				IsReply: true,
				Message: []string{"添加作业失败，错误原因：" + err.Error()},
			})
		}
		return c.JSON(http.StatusOK, MessageResponse{
			IsReply: true,
			Message: []string{"成功添加作业！作业科目：" + message.Param.Subject + "，作业内容：" + content},
		})
	}

	homework, err := model.FindHomeworkBySubject(message.Param.Subject)
	result := "查询到的作业内容如下：\n\n"
	for _, h := range homework {
		result += "科目：" + h.Subject + "\n"
		result += "内容：" + h.Content + "\n\n"
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
		ID:          "homework",
		Name:        "作业查询",
		Author:      "ligen131",
		Description: "作业查询：传入一个参数即科目，将返回这个科目的截止日期，如果你认为需要进行截止日期查询，但语句中没有给出科目，请传入语文。传入第二个参数即语句是否有意图添加作业，或语句中是否包含“添加”。传入第三个参数即作业内容。传入第四个参数即截止时间。",
		Prompt:      "作业查询",
		Params: []PluginParam{
			{
				Key:         "subject",
				Type:        "string",
				Description: "科目名",
			},
			{
				Key:         "isAddHomework",
				Type:        "string",
				Description: "语句是否有意图添加作业，或语句中是否包含“添加”",
			},
			{
				Key:         "Content",
				Type:        "string",
				Description: "作业内容",
			},
			{
				Key:         "Deadline",
				Type:        "string",
				Description: "截止时间",
			},
		},
		Format: []string{
			"${subject}作业什么时候截止",
			"${subject}作业截止日期",
			"${subject}作业截止时间",
			"${subject}作业截止",
			"${subject}作业什么时候交",
			"添加${subject}作业，内容为${Content}，${Deadline}截止",
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
			logs.Logs.Error("POST Plugin Center Register Endpoint failed", zap.Error(err))
		} else {
			logs.Logs.Error("POST Plugin Center Register Endpoint failed", zap.Int("StatusCode", resp.StatusCode), zap.Error(err))
		}
		return err
	}
	return nil
}

func main() {
	logs.InitLogs()

	// 读取配置文件
	config, err := YamlConfigLoad(path)
	if err != nil {
		panic(err)
	}

	err = model.Connect(config.Database)
	if err != nil {
		panic(err)
	}

	err = model.InitModel()
	if err != nil {
		panic(err)
	}

	err = Register(config.PluginCenterEndpoint, config.PluginEndpoint)
	if err != nil {
		logs.Logs.Error("Register failed", zap.Error(err))
	}
	// 每隔 1 分钟自动注册一次插件
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			err := Register(config.PluginCenterEndpoint, config.PluginEndpoint)
			if err != nil {
				logs.Logs.Error("Register failed", zap.Error(err))
			}
		}
	}()

	// 配置后端服务
	e := echo.New()
	e.POST("/", process)

	// 启动！
	e.Logger.Fatal(e.Start(config.Server.Hostname + ":" + fmt.Sprint(config.Server.Port)))
}
