package main

import (
	"../config"
	"../proxy-core"
	"../util"
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
	if err != nil {
		log.Fatalf("listen %s fail: %s", config.GlobalConfig.ServerPort, err)
	}
	util.PanicIfErr(err)
	log.Println("服务端监听端口:" + config.GlobalConfig.ServerPort)
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

func fetchServer(proxyPort string, client net.Conn) proxy_core.Server {
	server := portConnMap[proxyPort]
	if server.Server == nil {
		proxyListen, err := proxy_core.ListenServer(proxyPort)
		if err != nil {
			log.Println("服务端端口开启监听失败,端口:"+proxyPort, err)
			return proxy_core.Server{}
		}
		server = proxy_core.Server{proxyListen, 0, client, proxyPort}
		portConnMap[proxyPort] = server
	} else {
		//如果之前已经对这个端口发起过监听, 需要阻断之前得
		listen := portConnMap[proxyPort].Server
		v := portConnMap[proxyPort].V + 1
		server := proxy_core.Server{listen, v, client, proxyPort}
		portConnMap[proxyPort] = server
		time.Sleep(2 * time.Second)
	}
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
	return "代理的ip:" + p.Ip + " 代理的端口" + p.Port
}

var portConnMap = make(map[string]proxy_core.Server)

func firstConn(client net.Conn) ProxyAddress {
	buffer := make([]byte, 1024)
	//第一次连接进来我先要客户端先把代理端口传过来
	n, err := client.Read(buffer)
	var proxyAddr ProxyAddress
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
	//服务器对端口进行监听  主要监听访问层过来的
	/*if proxyPort == "8090" {
		proxyPort = "8091"
	}*/
	//兼容客户端重连
	server := fetchServer(proxyPort, client)
	if server.Server == nil {
		return
	}
	//****************************************
	//同一个端口的请求通过管道 来实现单线程
	connChan := make(chan proxy_core.Request, 100)
	go handlerConnChan(connChan, proxyPort)
	//****************************************
	server = portConnMap[proxyPort]
	go accept(server, connChan)
	for {
		if server.Expire(portConnMap[server.ProxyPort].V) {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func handlerConnChan(connChan chan proxy_core.Request, proxyPort string) {
	for {
		request := <-connChan
		/*_ = client.SetDeadline(time.Now().Add(5 * time.Second))
		_ = request.Conn.SetDeadline(time.Now().Add(5 * time.Second))*/
		log.Println(request.Conn.RemoteAddr())
		go proxy_core.ProxySwap(request.Conn, portConnMap[proxyPort].Client)
	}
}

func accept(server proxy_core.Server, connChan chan proxy_core.Request) {
	for {
		if server.Expire(portConnMap[server.ProxyPort].V) {
			return
		}
		proxyConn, err := server.Server.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(proxyConn)
		connChan <- proxy_core.Request{proxyConn, nil}
	}
}
