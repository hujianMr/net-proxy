package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	proxyStart(9300, 8090)
}
func proxyStart(fromport, toport int) {
	proxyaddr := fmt.Sprintf(":%d", fromport)
	proxylistener, err := net.Listen("tcp", proxyaddr)
	if err != nil {
		fmt.Println("Unable to listen on: %s, error: %s\n", proxyaddr, err.Error())
		os.Exit(1)
	}
	defer proxylistener.Close()

	for {
		proxyconn, err := proxylistener.Accept()
		if err != nil {
			fmt.Printf("Unable to accept a request, error: %s\n", err.Error())
			continue
		}

		// Read a header firstly in case you could have opportunity to check request
		// whether to decline or proceed the request
		buffer := make([]byte, 1024)
		n, err := proxyconn.Read(buffer)
		if err != nil {
			fmt.Printf("Unable to read from input, error: %s\n", err.Error())
			continue
		}

		// TODO
		// Your choice to make decision based on request header

		targetaddr := fmt.Sprintf("192.168.61.240:%d", toport)
		targetconn, err := net.Dial("tcp", targetaddr)
		if err != nil {
			fmt.Println("Unable to connect to: %s, error: %s\n", targetaddr, err.Error())
			proxyconn.Close()
			continue
		}

		n, err = targetconn.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Unable to write to output, error: %s\n", err.Error())
			proxyconn.Close()
			targetconn.Close()
			continue
		}

		go proxyRequest(proxyconn, targetconn)
		go proxyRequest(targetconn, proxyconn)
	}
}

// Forward all requests from r to w
func proxyRequest(r net.Conn, w net.Conn) {
	defer r.Close()
	defer w.Close()

	var buffer = make([]byte, 4096000)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			fmt.Printf("Unable to read from input, error: %s\n", err.Error())
			break
		}

		n, err = w.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Unable to write to output, error: %s\n", err.Error())
			break
		}
	}
}
