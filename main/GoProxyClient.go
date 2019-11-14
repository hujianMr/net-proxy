package main

import (
	"../config"
	"../util"
	"log"
	"net"
	"strings"
)

func main() {
	config.InitClientConfig()
	for _, proxyHost := range config.ClientProxyHosts {
		go handleProxyPort(strings.TrimSpace(proxyHost))
	}
	select {}
}

/**
  处理客户端代理的端口 端口跟阿里云开启的端口保持一致 当拨号阿里云端口的时候
  把端口上传到阿里云服务端，由服务端进行开启对外tcp端口 保持长连接
  服务端将外部请求路由到客户端
*/
func handleProxyPort(proxyHost string) {
	proxyPort := strings.Split(proxyHost, ":")[1]
	log.Println("客户端代理端口: " + proxyPort)
	serverUrl := config.ServerUrl
	//拨号连接服务端  建立长连接
	serverConn, err := net.Dial("tcp", serverUrl)
	if err != nil {
		log.Println("代理端口:" + proxyPort + " 拨号失败")
		return
	}
	// 把需要代理的内网ip:端口发送给
	_, err = serverConn.Write([]byte(proxyHost))
	if err != nil {
		log.Println("代理端口:" + proxyPort + " 写入端口失败")
		return
	}
	//接收到服务端返回代理请求的时候拨号 客户端代理端口 拨号代理端口
	proxyConn, err := net.Dial("tcp", proxyHost)
	if err != nil {
		log.Println(err)
		return
	}
	go util.ProxyRequestNotCloseConn(serverConn, proxyConn)
	go util.ProxyRequestNotCloseConn(proxyConn, serverConn)
	select {}
	/*for{
		var buffer = make([]byte, 4096000)
		//读取到服务端数据  代理客户端时跟服务端长连接的 这里监听每次服务端写过的数据
		n, err := serverConn.Read(buffer)
		if err != nil {
			continue
		}
		//把服务器数据写到代理端口服务
		n, err = proxyConn.Write(buffer[:n])
		if err != nil {
			continue
		}

	}*/
}
