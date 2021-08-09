package fttcp

import (
	"fmt"
	"net"
)

type FTTcpClient struct {
	conn net.Conn
}

func (ftTcpClient *FTTcpClient) Init(connect string) {
	c, err := net.Dial("tcp4", connect)
	if err != nil {
		fmt.Println(err)
		return
	}
	ftTcpClient.conn = c
}

func (ftTcpClient *FTTcpClient) Send(packet string) {
	fmt.Fprintf(ftTcpClient.conn, packet+"\n")
}

func (ftTcpClient *FTTcpClient) Close() {
	ftTcpClient.conn.Close()
}
