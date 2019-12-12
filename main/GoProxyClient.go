package main

import (
	"../config"
	"../proxy_core"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	//proxy_core.PrintWelcome()

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
	proxyPort, _ := strconv.Atoi(strings.Split(proxyHost, ":")[1])
	serverUrl := config.GlobalConfig.ServerUrl

	connChan := make(chan net.Conn)
	flagChan := make(chan bool)

	// 拨号
	go func(connCh chan net.Conn, flagCh chan bool) {
		for {
			select {
			case <-flagCh:
				go func(ch chan net.Conn) {
					conn := dial(serverUrl)
					if conn == nil {
						return
					}
					if requestConn(conn, proxyPort) {
						ch <- conn
						return
					}
					log.Println("Bridge connection interrupted,", proxyPort)
					_ = conn.Close()
					flagCh <- true
				}(connCh)
			default:
				// default
			}
		}
	}(connChan, flagChan)

	// 连接
	go func(connCh chan net.Conn, flagCh chan bool) {
		for {
			select {
			case cn := <-connCh:
				go func(conn net.Conn) {
					localConn := dial(proxyHost)
					if localConn == nil {
						_ = conn.Close()
						flagCh <- true // 通知创建连接
						return
					}
					flagCh <- true // 通知创建连接
					proxy_core.ProxySwap(localConn, conn)
				}(cn)
			default:
				// default
			}
		}
	}(connChan, flagChan)

	// 初始化连接
	flagChan <- true
}

func requestConn(conn net.Conn, proxyPort int) bool {
	port := int32(proxyPort)
	if !proxy_core.WritePort(conn, port) {
		return false
	}
	var resp int32
	if err := binary.Read(conn, binary.BigEndian, &resp); err != nil {
		return false
	}
	log.Println("proxy service", port)
	return true
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
		time.Sleep(5 * time.Second)
	}
}
