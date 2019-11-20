package config

import (
	"../util"
	"github.com/Unknwon/goconfig/goconfig-master"
	"strings"
)

func InitServerConfig() {
	cfg, err := goconfig.LoadConfigFile("config.ini")
	util.PanicIfErrMsg(err, "找不到配置文件 config.ini")
	GlobalConfig.ServerPort, err = cfg.GetValue("server", "port")
	util.PanicIfErrMsg(err, "serverPort参数配置为空")
	initPortMapping()
}

var GlobalConfig Config

type Config struct {
	ClientProxyHosts []string
	ServerUrl        string
	ServerPort       string
}

func InitClientConfig() {
	cfg, err := goconfig.LoadConfigFile("config.ini")
	util.PanicIfErrMsg(err, "找不到配置文件 config.ini")

	proxyHosts, err := cfg.GetValue("client", "proxyHosts")
	util.PanicIfErrMsg(err, "proxyHosts参数配置为空")

	GlobalConfig.ClientProxyHosts = strings.Split(proxyHosts, ",")
	GlobalConfig.ServerUrl, err = cfg.GetValue("client", "serverUrl")
	util.PanicIfErrMsg(err, "serverUrl参数配置为空")
}

func GetMappingPort(port string) string {
	realPort := portMapping[port]
	if realPort == "" {
		realPort = port
	}
	return realPort
}

var portMapping = make(map[string]string)

func initPortMapping() {
	portMapping["22"] = "10022"
	portMapping["80"] = "10080"
}
