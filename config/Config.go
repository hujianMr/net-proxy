package config

import (
	"../util"
	"github.com/Unknwon/goconfig/goconfig-master"
	"strings"
)

var ServerPort string

func InitServerConfig() {
	cfg, err := goconfig.LoadConfigFile("config.ini")
	util.PanicIfErrMsg(err, "找不到配置文件 config.ini")
	ServerPort, err = cfg.GetValue("server", "port")
	util.PanicIfErrMsg(err, "serverPort参数配置为空")
}

var ClientProxyHosts []string
var ServerUrl string

func InitClientConfig() {
	cfg, err := goconfig.LoadConfigFile("config.ini")
	util.PanicIfErrMsg(err, "找不到配置文件 config.ini")

	proxyHosts, err := cfg.GetValue("client", "proxyHosts")
	util.PanicIfErrMsg(err, "proxyHosts参数配置为空")

	ClientProxyHosts = strings.Split(proxyHosts, ",")
	ServerUrl, err = cfg.GetValue("client", "serverUrl")
	util.PanicIfErrMsg(err, "serverUrl参数配置为空")

}
