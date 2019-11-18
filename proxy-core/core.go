package proxy_core

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Request struct {
	Conn net.Conn
	Buff []byte
}

type Server struct {
	Server net.Listener
	V      int
	Client net.Conn
}

func ListenServer(address string) (net.Listener, error) {
	log.Println("监听地址：" + address)
	server, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func ProxySwap(proxyConn net.Conn, client net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go ConnCopy(proxyConn, client, &wg)
	go ConnCopy(client, proxyConn, &wg)
	wg.Wait()
	log.Println("conn1 = [" + proxyConn.LocalAddr().String() + "], conn2 = [" + client.RemoteAddr().String() + "] iocopy读写完成")
}
func ConnCopy(conn1 net.Conn, conn2 net.Conn, wg *sync.WaitGroup) {
	_, err := io.Copy(conn1, conn2)
	if err != nil {
		log.Println("conn1 = ["+conn1.LocalAddr().String()+"], conn2 = ["+conn2.RemoteAddr().String()+"] iocopy失败", err)
	}
	log.Println("[←]", "close the connect at local:["+conn1.LocalAddr().String()+"] and remote:["+conn1.RemoteAddr().String()+"]")
	conn1.Close()
	wg.Done()
}

func PrintWelcome() {
	fmt.Println("+----------------------------------------------------------------+")
	fmt.Println("| welcome use ho-huj-net-proxy Version1.0                        |")
	fmt.Println("| author Ruchsky at 2019-11-14                                   |")
	fmt.Println("| gitee home page ->   https://gitee.com/ruchsky                 |")
	fmt.Println("| github home page ->  https://github.com/hujianMr               |")
	fmt.Println("+----------------------------------------------------------------+")
	fmt.Print("start..")
	i := 0
	for {
		fmt.Print(">>>>")
		i++
		time.Sleep(time.Second)
		if i >= 10 {
			break
		}
	}
	fmt.Println()
	fmt.Println("start success")
	time.Sleep(time.Second)
}
