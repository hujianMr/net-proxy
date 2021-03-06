package proxy_core

import (
	"encoding/binary"
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
	Server    net.Listener
	V         int
	Client    net.Conn
	ProxyPort int32
}

func (s *Server) IncrCycle(client net.Conn) Server {
	s.V++
	s.Client = client
	return Server{
		Server:    s.Server,
		V:         s.V,
		Client:    s.Client,
		ProxyPort: s.ProxyPort,
	}
}

func (s *Server) Expire(V int) bool {
	if s.V < V {
		return true
	}
	return false
}

// 关闭连接
func Close(cons ...net.Conn) {
	for _, conn := range cons {
		_ = conn.Close()
	}
}

func ListenServer(proxyPort int32) (net.Listener, error) {
	address := fmt.Sprintf("0.0.0.0:%d", proxyPort)
	log.Println("listen addr ：" + address)
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
	//log.Println("conn1 = [" + proxyConn.LocalAddr().String() + "], conn2 = [" + client.RemoteAddr().String() + "] iocopy读写完成")
}
func ConnCopy(conn1 net.Conn, conn2 net.Conn, wg *sync.WaitGroup) {
	_, err := io.Copy(conn1, conn2)
	if err != nil {
		// 连接中断，不打印日志
		//log.Println("conn1 = ["+conn1.LocalAddr().String()+"], conn2 = ["+conn2.RemoteAddr().String()+"] iocopy失败", err)
	}
	//log.Println("[←]", "close the connect at local:["+conn1.LocalAddr().String()+"] and remote:["+conn1.RemoteAddr().String()+"]")
	_ = conn1.Close()
	wg.Done()
}

func WritePort(conn net.Conn, port int32) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		log.Println("Fail to set write deadline", err)
		return false
	}
	if err := binary.Write(conn, binary.BigEndian, port); err != nil {
		return false
	}
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		log.Println("Fail to clear write deadline", err)
		return false
	}
	return true
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
