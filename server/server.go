package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	tcp "../fttcp"
	udp "../ftudp"
)

type FTServer struct {
	filePath  string
	file      *os.File
	log       *os.File
	tags      string
	tcpClient *tcp.FTTcpClient
	connected bool
	chan_prog chan bool
}

var GClient_Address string
var GFullFilePath string
var fileSize int64
var fileSizeReceived int64

func main() {
	GClient_Address = os.Args[1]
	GFullFilePath = os.Args[2]

	ftServer := &FTServer{filePath: GFullFilePath}
	ftServer.init()
	fileSize = 0
	fileSizeReceived = 0
	go udp.InitUDPServer("localhost", 1234, ftServer.handleUDPServerAction, ftServer.handleError)
	go tcp.InitTCPServer(":12341", func(arg string) {

		if ftServer.connected == false {
			ftServer.tcpClient.Init(GClient_Address + ":4000")
			ftServer.connected = true
		}
		fmt.Println(arg)
		if strings.Contains(arg, "FileSize:") {
			fileInfo := strings.Split(arg, ":")
			if len(fileInfo) > 1 {
				fileSize, _ = strconv.ParseInt(fileInfo[1], 10, 64)
			}
			fmt.Printf("Total size of file to be copied is %v \n", fileSize)
			ftServer.tcpClient.Send("START_SENDING")
		} else if arg == "PASS" {
			ftServer.tcpClient.Send(ftServer.tags)
			ftServer.tags = ""
		} else if arg == "EXIT" {
			ftServer.tcpClient.Close()
			fmt.Println("File Copied successfully")
			ftServer.chan_prog <- true
		}
	})

	<-ftServer.chan_prog
}

func (ftServer *FTServer) init() {
	ftServer.log, _ = os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	ftServer.file, _ = os.Create(ftServer.filePath)
	if ftServer.file == nil {
		log.Fatalln("Unable to create file")
	}
	if ftServer.log == nil {
		log.Fatalln("Unable to create log file")
	}
	ftServer.tcpClient = &tcp.FTTcpClient{}
	ftServer.connected = false

	ftServer.chan_prog = make(chan bool)

	ftServer.tags = ""
	log.SetOutput(ftServer.log)
}

func (ftServer *FTServer) writeFile(data []byte, offset int) {
	MAX_LEN := 10 * 1024
	_, errWrite := ftServer.file.WriteAt(data, int64((offset)*MAX_LEN))

	if errWrite != nil {
		fmt.Println("error in writing to file")
	}
}

func (ftServer *FTServer) handleUDPServerAction(data []byte) {
	offset := bytesToInt(data[0:4])
	log.Println(strconv.Itoa(int(offset)))
	ftServer.tags += strconv.Itoa(int(offset)) + ","
	ftServer.writeFile(data[4:], int(offset))
	fileSizeReceived += int64(len(data[4:]))

	if fileSizeReceived < fileSize {
		progress := float64((float32(fileSizeReceived) / float32(fileSize)) * 100)
		fmt.Printf("File received %v\n", math.Floor(progress*100)/100)
	}
}

func (ftServer *FTServer) handleError(err string) {
	ftServer.tcpClient.Close()
}

func bytesToInt(arrByte []byte) uint32 {
	data := binary.BigEndian.Uint32(arrByte)
	return data
}
