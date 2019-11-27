package main

import (
	"../config"
	"../proxy-core"
	"bytes"
	"encoding/binary"
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

		// 把需要代理的内网ip:端口发送给
		if err := write(serverConn, proxyHost); err != nil {
			log.Printf("first write failure > port = %s", proxyHost)
			proxy_core.Close(serverConn, proxyConn)
			time.Sleep(2 * time.Second)
			continue
		}
		//拨号代理端口
		proxyConn = dial(proxyHost)
		proxy_core.ProxySwap(serverConn, proxyConn)

	}
}

/**
  处理报文头 固定长度
*/
func write(conn net.Conn, content string) error {
	_contentbytes := []byte(content)
	len := int32(len(_contentbytes))
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, len)
	_lenBytes := bytesBuffer.Bytes()
	var buffer bytes.Buffer
	buffer.Write(_lenBytes)
	buffer.Write(_contentbytes)
	_, err := conn.Write(buffer.Bytes())
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
