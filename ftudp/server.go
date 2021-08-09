package ftudp

import (
	"fmt"
	"log"
	"net"
)

func InitUDPServer(host string, port int, callbackData func(data []byte), callBackError func(err string)) {

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	serConn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	handleRequest(serConn, callbackData, callBackError)
}

func handleRequest(ser *net.UDPConn, callbackData func(data []byte), callBackError func(err string)) {

	MAX_LEN := 10*1024 + 4
	buffer := make([]byte, MAX_LEN)

	for {
		bytesRead, _, err := ser.ReadFromUDP(buffer)

		if err != nil {
			log.Println("Reading from UDP channel fails")
			callBackError("Reading from UDP channel fails")
			continue
		}

		callbackData(buffer[0:bytesRead])
	}
}
