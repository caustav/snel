package ftudp

import (
	"log"
	"net"
)

type UdpClient struct {
	conn *net.UDPConn
}

func (udpClient *UdpClient) Init(address string) {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatalln("Unable to resolve address")
	}

	udpClient.conn, err = net.DialUDP("udp4", nil, s)

	if err != nil {
		log.Fatalln("Unable to dial address")
	}
}

func (udpClient *UdpClient) Send(data []byte) int {

	lenWrite, errWrite := udpClient.conn.Write(data)

	if errWrite != nil {
		log.Println("Error sending data to server")
	}

	return lenWrite
}

func (udpClient *UdpClient) Close() {
	udpClient.conn.Close()
}
