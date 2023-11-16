package controllers

import (
	"bytes"
	"carrota-plugin-center/model"
	"carrota-plugin-center/shared/service"
	"carrota-plugin-center/utils"
	"carrota-plugin-center/utils/logs"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func wrapAndSendMessage(originMessage model.MessageInfo, message []string) error {
	// 提交 Wrapper
	wrapperRequest := model.PostWrapperRequest{
		Agent:          originMessage.Agent,
		GroupID:        originMessage.GroupID,
		GroupName:      originMessage.GroupName,
		UserID:         originMessage.UserID,
		UserName:       originMessage.UserName,
		Time:           time.Now().Unix(),
		Message:        originMessage.Message,
		OriginResponse: message,
	}
	jsonStr, _ := json.Marshal(wrapperRequest)
	req, _ := http.NewRequest("POST", service.WrapperEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		logs.Error("POST Wrapper endpoint failed", zap.Int("statusCode", resp.StatusCode), zap.Error(err))
		return err
	}

	wrapperResponse := model.PostWrapperResponse{}
	err = json.NewDecoder(resp.Body).Decode(&wrapperResponse)
	logs.Debug("wrapperResponse", zap.Any("wrapperResponse", wrapperResponse))
	if err != nil {
		logs.Error("Decode wrapperResponse failed", zap.Error(err))
		return err
	}
	resp.Body.Close()

	// 提交 Agent 发送信息
	jsonStr, _ = json.Marshal(model.MessageSendRequest{
		Agent:     originMessage.Agent,
		MessageID: originMessage.MessageID,
		GroupID:   originMessage.GroupID,
		UserID:    originMessage.UserID,
		Message:   wrapperResponse.Response,
	})
	req, _ = http.NewRequest("POST", service.AgentEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		logs.Error("POST Agent endpoint failed", zap.Int("statusCode", resp.StatusCode), zap.Error(err))
		return err
	}

	return nil
}

func processUserMessage(message model.MessageInfo) error {
	// 提交 Parser
	jsonStr, _ := json.Marshal(message)
	req, _ := http.NewRequest("POST", service.ParserEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		logs.Error("POST Parser endpoint failed", zap.Int("statusCode", resp.StatusCode), zap.Error(err))
		return err
	}

	parserResponse := model.ParserResponse{}
	err = json.NewDecoder(resp.Body).Decode(&parserResponse)
	if err != nil {
		logs.Error("Decode parserResponse failed", zap.Error(err))
		return err
	}
	logs.Debug("parserResponse", zap.Any("parserResponse", parserResponse))
	resp.Body.Close()

	// 提交 Plugin
	messageReply := model.MessageReply{}
	for _, parserPlugin := range parserResponse.Plugin {
		plugin, err := model.FindPluginById(parserPlugin.ID)
		if err != nil {
			continue
		}

		pluginStr, _ := json.Marshal(model.PostPluginRequest{
			Agent:     message.Agent,
			MessageID: message.MessageID,
			GroupID:   message.GroupID,
			GroupName: message.GroupName,
			UserID:    message.UserID,
			UserName:  message.UserName,
			Time:      message.Time,
			Message:   message.Message,
			IsMention: message.IsMention,
			Param:     parserPlugin.Param,
		})
		req, _ := http.NewRequest("POST", plugin.Url, bytes.NewBuffer(pluginStr))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		var resp *http.Response
		for i := 0; i < utils.FailedAttempts; i++ {
			resp, err = client.Do(req)
			if err == nil && resp.StatusCode != 200 {
				break
			}
		}

		if err != nil || resp.StatusCode != 200 {
			model.DeletePluginById(plugin.ID)
			logs.Warn("POST Plugin endpoint failed", zap.String("name", plugin.Name), zap.String("url", plugin.Url), zap.Int("statusCode", resp.StatusCode), zap.Error(err))
			continue
		}

		pluginResponse := model.MessageReply{}
		err = json.NewDecoder(resp.Body).Decode(&pluginResponse)
		if err != nil {
			logs.Error("Decode pluginResponse failed", zap.Error(err))
			return err
		}
		resp.Body.Close()

		messageReply.IsReply = messageReply.IsReply || pluginResponse.IsReply
		messageReply.Message = append(messageReply.Message, pluginResponse.Message...)
	}
	if messageReply.IsReply || true {
		err = wrapAndSendMessage(message, messageReply.Message)
		if err != nil {
			return err
		}
	}
	return nil
}

func MessagePOST(c echo.Context) error {
	logs.Debug("GET /message")

	message := model.MessageInfo{}
	_ok, err := Bind(c, &message)
	if !_ok {
		return err
	}

	go processUserMessage(message)

	return ResponseOK(c, "ok")
}

func MessageSendPOST(c echo.Context) error {
	logs.Debug("GET /message")

	message := model.MessageSendRequest{}
	_ok, err := Bind(c, &message)
	if !_ok {
		return err
	}

	err = wrapAndSendMessage(model.MessageInfo{
		MessageID: message.MessageID,
		Agent:     message.Agent,
		GroupID:   message.GroupID,
		UserID:    message.UserID,
	}, message.Message)
	if err != nil {
		return ResponseInternalServerError(c, "Send message failed", err)
	}

	return ResponseOK(c, "ok")
}
