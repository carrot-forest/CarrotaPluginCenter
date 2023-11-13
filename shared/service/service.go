package service

type CarrotaServiceConfig struct {
	AgentEndpoint   string `config:"agent-endpoint"`
	ParserEndpoint  string `config:"parser-endpoint"`
	WrapperEndpoint string `config:"wrapper-endpoint"`
}

var AgentEndpoint string
var ParserEndpoint string
var WrapperEndpoint string

func CarrotaServiceConfigInit(c CarrotaServiceConfig) error {
	AgentEndpoint = c.AgentEndpoint
	ParserEndpoint = c.ParserEndpoint
	WrapperEndpoint = c.WrapperEndpoint
	return nil
}
