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

func MessagePOST(c echo.Context) error {
	logs.Debug("GET /message")

	message := model.MessageInfo{}
	_ok, err := Bind(c, &message)
	if !_ok {
		return err
	}

	// 提交 Parser
	jsonStr, _ := json.Marshal(message)
	req, _ := http.NewRequest("POST", service.ParserEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ResponseInternalServerError(c, "POST Parser endpoint failed", err)
	}

	parserResponse := model.ParserResponse{}
	err = json.NewDecoder(resp.Body).Decode(&parserResponse)
	if err != nil {
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
			GroupID:   message.GroupID,
			GroupName: message.GroupName,
			UserID:    message.UserID,
			UserName:  message.UserName,
			Time:      message.Time,
			Message:   message.Message,
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
			continue
		}

		pluginResponse := model.MessageReply{}
		err = json.NewDecoder(resp.Body).Decode(&pluginResponse)
		if err != nil {
			return err
		}
		resp.Body.Close()

		messageReply.IsReply = messageReply.IsReply || pluginResponse.IsReply
		messageReply.Message = append(messageReply.Message, pluginResponse.Message...)
	}

	// 提交 Wrapper
	jsonStr, _ = json.Marshal(model.PostWrapperRequest{
		Agent:          message.Agent,
		GroupID:        message.GroupID,
		GroupName:      message.GroupName,
		UserID:         message.UserID,
		UserName:       message.UserName,
		Time:           message.Time,
		Message:        message.Message,
		OriginResponse: messageReply.Message,
	})
	req, _ = http.NewRequest("POST", service.WrapperEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ResponseInternalServerError(c, "POST Wrapper endpoint failed", err)
	}

	wrapperResponse := model.PostWrapperResponse{}
	err = json.NewDecoder(resp.Body).Decode(&wrapperResponse)
	logs.Debug("wrapperResponse", zap.Any("wrapperResponse", wrapperResponse))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return ResponseOK(c, model.MessageReply{
		IsReply: messageReply.IsReply,
		Message: wrapperResponse.Response,
	})
}

func MessageSendPOST(c echo.Context) error {
	logs.Debug("GET /message")

	message := model.MessageSendRequest{}
	_ok, err := Bind(c, &message)
	if !_ok {
		return err
	}

	// 提交 Wrapper
	wrapperRequest := model.PostWrapperRequest{
		Agent:          message.Agent,
		Time:           time.Now().Unix(),
		OriginResponse: message.Message,
	}
	if message.IsPrivate {
		wrapperRequest.UserID = message.To
	} else {
		wrapperRequest.GroupID = message.To
	}
	jsonStr, _ := json.Marshal(wrapperRequest)
	req, _ := http.NewRequest("POST", service.WrapperEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ResponseInternalServerError(c, "POST Wrapper endpoint failed", err)
	}

	wrapperResponse := model.PostWrapperResponse{}
	err = json.NewDecoder(resp.Body).Decode(&wrapperResponse)
	logs.Debug("wrapperResponse", zap.Any("wrapperResponse", wrapperResponse))
	if err != nil {
		return err
	}
	resp.Body.Close()

	// 提交 Agent 发送信息
	jsonStr, _ = json.Marshal(model.MessageSendRequest{
		Agent:     message.Agent,
		IsPrivate: message.IsPrivate,
		To:        message.To,
		Message:   wrapperResponse.Response,
	})
	req, _ = http.NewRequest("POST", service.AgentEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ResponseInternalServerError(c, "POST Agent endpoint failed", err)
	}

	return ResponseOK(c, "ok")
}
