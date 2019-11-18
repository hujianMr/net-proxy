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
	addr := "0.0.0.0:" + config.ServerPort
	//监听服务端端口
	serverListen, err := proxy_core.ListenServer(addr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	}
	util.PanicIfErr(err)
	log.Println("服务端监听端口:" + config.ServerPort)
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

var portConnMap = make(map[string]proxy_core.Server)

func handleClient(client net.Conn) {
	buffer := make([]byte, 1024)
	//第一次连接进来我先要客户端先把代理端口传过来
	n, err := client.Read(buffer)
	if err != nil {
		fmt.Printf("Unable to read from input, error: %s\n", err.Error())
		return
	}
	proxyHost := string(buffer[:n])
	proxyIp := strings.Split(proxyHost, ":")[0]
	proxyPort := strings.Split(proxyHost, ":")[1]
	log.Println("代理的ip:" + proxyIp + " 代理的端口" + proxyPort)

	//服务器对端口进行监听  主要监听访问层过来的
	//这个8091端口是我自己测试用的  因为开发的时候用的同一台机器
	/*if proxyPort == "8090" {
		proxyPort = "8091"
	}*/
	//兼容客户端重连
	server := portConnMap[proxyPort]
	if server.Server == nil {
		addr := "0.0.0.0:" + proxyPort
		proxyListen, err := proxy_core.ListenServer(addr)
		if err != nil {
			log.Println("服务端端口开启监听失败,端口:"+proxyPort, err)
			return
		}
		portConnMap[proxyPort] = proxy_core.Server{proxyListen, 0, client}
	} else {
		//如果之前已经对这个端口发起过监听, 需要阻断之前得
		listen := portConnMap[proxyPort].Server
		v := portConnMap[proxyPort].V + 1
		server := proxy_core.Server{listen, v, client}
		portConnMap[proxyPort] = server
	}
	//****************************************
	//同一个端口的请求通过管道 来实现单线程
	connChan := make(chan proxy_core.Request, 100)
	go handlerConnChan(connChan, proxyPort)
	//****************************************
	go accept(portConnMap[proxyPort].Server, portConnMap[proxyPort].Client, connChan)
	server = portConnMap[proxyPort]
	for {
		if server.V < portConnMap[proxyPort].V {
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
		proxy_core.ProxySwap(request.Conn, portConnMap[proxyPort].Client)
		//request.Conn.Close()
	}
}

func accept(listener net.Listener, client net.Conn, connChan chan proxy_core.Request) {
	for {
		proxyConn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(proxyConn)
		connChan <- proxy_core.Request{proxyConn, nil}
	}
}
