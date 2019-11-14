package util

import (
	"fmt"
	"net"
)

/**
  io 复转 这里不关闭资源
*/
func ProxyRequest(formConn net.Conn, toConn net.Conn) {
	defer formConn.Close()
	defer toConn.Close()
	var buffer = make([]byte, 4096000)
	for {
		n, err := formConn.Read(buffer)
		if err != nil {
			fmt.Printf("服务端读取代理客户端数据错误, error: %s\n", err.Error())
			break
		}

		n, err = toConn.Write(buffer[:n])
		if err != nil {
			fmt.Printf("服务端写数据到代理客户端错误, error: %s\n", err.Error())
			break
		}
	}
}
