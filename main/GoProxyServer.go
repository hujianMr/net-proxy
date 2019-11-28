package main

import (
	"../config"
	"../proxy-core"
	"../util"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func main() {

	proxy_core.PrintWelcome()

	config.InitServerConfig()
	//监听服务端端口
	serverListen, err := proxy_core.ListenServer(config.GlobalConfig.ServerPort)
	util.PanicIfErr(err)
	log.Printf("server listen port = %s\n:", config.GlobalConfig.ServerPort)
	for {
		client, err := serverListen.Accept()
		if err != nil {
			continue
		}
		log.Println(client.RemoteAddr())
		//接收客户端需要路由的端口
		go handleClient(client)
	}
}

func _pakServer(server net.Listener, v int, client net.Conn, proxyPort string) proxy_core.Server {
	return proxy_core.Server{
		Server:    server,
		V:         v,
		Client:    client,
		ProxyPort: proxyPort,
	}
}

func fetchServer(proxyPort string, client net.Conn) proxy_core.Server {
	server := portConnMap[proxyPort]
	if server.Server == nil {
		proxyListen, err := proxy_core.ListenServer(proxyPort)
		if err != nil {
			log.Printf("server port listen failure, port = %s, errs = %s\n", proxyPort, err.Error())
			return proxy_core.Server{}
		}
		server = _pakServer(proxyListen, 0, client, proxyPort)
		portConnMap[proxyPort] = server
	}
	//替换新的结构体 新的结构体周期增加
	portConnMap[proxyPort] = server.IncrCycle(client)
	return server
}

type ProxyAddress struct {
	Ip   string
	Port string
}

func (p *ProxyAddress) convert(proxyHost string) {
	p.Ip = strings.Split(proxyHost, ":")[0]
	p.Port = config.GetMappingPort(strings.Split(proxyHost, ":")[1])
}

func (p *ProxyAddress) toString() string {
	return fmt.Sprintf("proxy id = %s proxy port = %s", p.Ip, p.Port)
}

var portConnMap = make(map[string]proxy_core.Server)

func firstConn(client net.Conn) ProxyAddress {
	var proxyAddr ProxyAddress
	buffer := make([]byte, 4)
	//第一次连接进来我先要客户端先把代理端口传过来
	n, err := client.Read(buffer)
	if err != nil {
		fmt.Printf("Unable to read from input, error: %s\n", err.Error())
		return proxyAddr
	}
	//读取报文长度
	var len int32
	err = binary.Read(bytes.NewBuffer(buffer[:n]), binary.BigEndian, &len)
	if err != nil {
		return proxyAddr
	}
	buffer = make([]byte, len)
	n, err = client.Read(buffer)
	if err != nil {
		fmt.Printf("Unable to read from input, error: %s\n", err.Error())
		return proxyAddr
	}
	proxyHost := string(buffer[:n])
	proxyAddr.convert(proxyHost)
	log.Println(proxyAddr.toString())
	return proxyAddr
}

func handleClient(client net.Conn) {
	proxyAddr := firstConn(client)
	if proxyAddr.Port == "" {
		return
	}
	proxyPort := proxyAddr.Port
	//兼容客户端重连
	server := fetchServer(proxyPort, client)
	if server.Server == nil {
		return
	}
	//希望所有的请求进入管道
	/*connChan := make(chan proxy_core.Request, 100)
	go handlerConnChan(connChan, proxyPort)*/

	server = portConnMap[proxyPort]
	go accept(server, callback, proxyPort)
	for {
		//当前版本号不匹配
		if server.Expire(portConnMap[server.ProxyPort].V) {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func callback(src net.Conn, dest net.Conn) {
	proxy_core.ProxyIoBind(&src, dest)
}

/*func handlerConnChan(connChan chan proxy_core.Request, proxyPort string) {
	for {
		request := <-connChan
		log.Println(request.Conn.RemoteAddr())
		go proxy_core.ProxySwap(request.Conn, portConnMap[proxyPort].Client)
	}
}*/

func accept(server proxy_core.Server, fn func(src net.Conn, dest net.Conn), proxyPort string) {
	for {
		if server.Expire(portConnMap[server.ProxyPort].V) {
			return
		}
		var srcConn net.Conn
		srcConn, err := server.Server.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			fn(srcConn, portConnMap[proxyPort].Client)
		}()
		//connChan <- proxy_core.Request{proxyConn, nil}
	}
}
