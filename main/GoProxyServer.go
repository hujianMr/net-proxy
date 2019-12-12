package main

import (
	"../config"
	"../proxy_core"
	"../util"
	"encoding/binary"
	"log"
	"net"
	"strconv"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	//proxy_core.PrintWelcome()

	config.InitServerConfig()
	//监听服务端端口
	serverPort, _ := strconv.Atoi(config.GlobalConfig.ServerPort)
	serverPort32 := int32(serverPort)
	serverListen, err := proxy_core.ListenServer(serverPort32)
	util.PanicIfErr(err)
	log.Printf("server listen port = %s\n", config.GlobalConfig.ServerPort)
	for {
		client, err := serverListen.Accept()
		if err == nil {
			log.Println(client.RemoteAddr())
			go handleClient(client)
		}
	}
}

func _pakServer(server net.Listener, v int, client net.Conn, proxyPort int32) proxy_core.Server {
	return proxy_core.Server{
		Server:    server,
		V:         v,
		Client:    client,
		ProxyPort: proxyPort,
	}
}

// TODO 需要加锁保证并发问题
func fetchServer(proxyPort int32, client net.Conn) proxy_core.Server {
	server := serverMap[proxyPort]
	if server.Server == nil {
		proxyListen, err := proxy_core.ListenServer(proxyPort)
		if err != nil {
			log.Printf("server port listen failure, port = %d, errs = %s\n", proxyPort, err.Error())
			return proxy_core.Server{}
		}
		server = _pakServer(proxyListen, 0, client, proxyPort)
		serverMap[proxyPort] = server
	}
	//替换新的结构体 新的结构体周期增加
	serverMap[proxyPort] = server.IncrCycle(client)
	return server
}

var serverMap = make(map[int32]proxy_core.Server)

func firstConn(client net.Conn) (realPort int32, ok bool) {
	var port int32
	if err := binary.Read(client, binary.BigEndian, &port); err != nil {
		log.Println("Fail to handle first conn", err)
		return
	}
	// get mapping port
	realPort = config.GetRealPort(port)
	log.Printf("first conn, port mapping %d -> %d\n", port, realPort)
	return realPort, true
}

func handleClient(client net.Conn) {
	proxyPort, ok := firstConn(client)
	if !ok {
		return
	}
	server := fetchServer(proxyPort, client)
	if server.Server == nil {
		return
	}
	server = serverMap[proxyPort]
	srcConn, err := serverMap[proxyPort].Server.Accept()
	if err != nil {
		_ = client.Close()
		return
	}
	// 发送连接通知, 注意 port 必须是 int32
	if proxy_core.WritePort(client, proxyPort) {
		proxy_core.ProxySwap(srcConn, serverMap[proxyPort].Client)
	} else {
		_ = client.Close()
		_ = srcConn.Close()
		log.Println("fail to conn", err)
	}
}
