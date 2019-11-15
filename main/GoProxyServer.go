package main

import (
	"../config"
	"../proxy-core"
	"../util"
	"fmt"
	"log"
	"net"
	"strings"
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
	if proxyPort == "8090" {
		proxyPort = "8091"
	}
	addr := "0.0.0.0:" + proxyPort
	proxyListen, err := proxy_core.ListenServer(addr)
	if err != nil {
		log.Println("服务端端口开启监听失败,端口:"+proxyPort, err)
		return
	}
	for {
		proxyConn, err := proxyListen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		/*buffer := make([]byte, 1024)
		n, err := proxyConn.Read(buffer)
		if err != nil || n <= 0 {
			fmt.Printf("服务端读取访问客户端》》数据错误, error: %s\n", err.Error())
			continue
		}*/
		//log....
		proxy_core.ProxySwap(proxyConn, client)
	}
}
