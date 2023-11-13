package model

type MessageInfo struct {
	Agent     string `json:"agent"`
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Time      int64  `json:"time"`
	Message   string `json:"message"`
}

type MessageReply struct {
	IsReply bool     `json:"is_reply"`
	Message []string `json:"message"`
}

type ParserPluginInfo struct {
	ID    string      `json:"id"`
	Param interface{} `json:"param"`
}

type ParserResponse struct {
	Plugin []ParserPluginInfo `json:"plugin"`
}

type PostPluginRequest struct {
	Agent     string      `json:"agent"`
	GroupID   string      `json:"group_id"`
	GroupName string      `json:"group_name"`
	UserID    string      `json:"user_id"`
	UserName  string      `json:"user_name"`
	Time      int64       `json:"time"`
	Message   string      `json:"message"`
	Param     interface{} `json:"param"`
}

type PostWrapperRequest struct {
	Agent          string   `json:"agent"`
	GroupID        string   `json:"group_id"`
	GroupName      string   `json:"group_name"`
	UserID         string   `json:"user_id"`
	UserName       string   `json:"user_name"`
	Time           int64    `json:"time"`
	Message        string   `json:"message"`
	OriginResponse []string `json:"origin_response"`
}

type PostWrapperResponse struct {
	Response []string `json:"response"`
}

type MessageSendRequest struct {
	Agent     string   `json:"agent"`
	IsPrivate bool     `json:"is_private"`
	To        string   `json:"to"`
	Message   []string `json:"message"`
}
