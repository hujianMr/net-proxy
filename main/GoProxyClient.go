package main

import (
	"../config"
	"../proxy-core"
	"log"
	"net"
	"strings"
	"time"
)

func main() {

	proxy_core.PrintWelcome()

	config.InitClientConfig()
	for _, proxyHost := range config.GlobalConfig.ClientProxyHosts {
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
	log.Printf("proxy client port = %s\n ", proxyPort)
	var proxyConn, serverConn net.Conn
	serverUrl := config.GlobalConfig.ServerUrl
	i := 0
	for {
		i++
		log.Printf("connection times=(%d) , proxyPort = %s\n", i, proxyPort)
		//拨号连接服务端
		serverConn = dial(serverUrl)
		//拨号代理端口
		proxyConn = dial(proxyHost)
		// 把需要代理的内网ip:端口发送给
		if err := write(serverConn, []byte(proxyHost)); err != nil {
			log.Printf("first write failure > port = %s", proxyHost)
			proxy_core.Close(serverConn, proxyConn)
			time.Sleep(2 * time.Second)
			continue
		}
		proxy_core.ProxySwap(serverConn, proxyConn)

	}
}

func write(conn net.Conn, bytes []byte) error {
	_, err := conn.Write(bytes)
	return err
}

func dial(dialUrl string) net.Conn {
	log.Printf("client start dial > url = %s", dialUrl)
	for {
		conn, err := net.Dial("tcp", dialUrl)
		if err == nil {
			log.Printf("dial success > url = %s", dialUrl)
			return conn
		}
		log.Printf("dial failure > url = %s, errmsg = %s\n", dialUrl, err.Error())
		time.Sleep(3 * time.Second)
	}
}
