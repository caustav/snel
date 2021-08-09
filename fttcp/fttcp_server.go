package fttcp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

var Gl net.Listener

func InitTCPServer(port string, callback func(arg string)) {
	l, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println(err)
		return
	}
	Gl = l
	defer l.Close()
	fmt.Println("TCP Server Listening")
	for {
		c, err := l.Accept()
		defer c.Close()

		if err != nil {
			fmt.Print("InitTCPServer ")
			fmt.Println(err)
			return
		}
		go handleConnection(c, callback)
	}
}

func handleConnection(c net.Conn, callback func(arg string)) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		dataUnformatted, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Printf("Closing client connection %v", err)
			return
		}
		data := strings.TrimSpace(string(dataUnformatted))
		callback(data)
	}
}

func Close() {
	Gl.Close()
}
