package logs

import "go.uber.org/zap"

var Logs *zap.Logger

func InitLogs() {
	// logger 初始化
	Logs, _ = zap.NewDevelopment()

}
