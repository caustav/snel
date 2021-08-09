package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	tcp "../fttcp"
	udp "../ftudp"
)

type FTClient struct {
	tags      map[string]bool
	file      *FileHandle
	udpClient *udp.UdpClient
	tcpClient *tcp.FTTcpClient
	chan_prog chan bool
}

type FileHandle struct {
	filePath string
	file     *os.File
	size     int64
}

var GServer_Address string
var GFullFilePath string

func main() {

	GServer_Address = os.Args[1]
	GFullFilePath = os.Args[2]

	file := &FileHandle{filePath: GFullFilePath}
	file.Init()

	ftClient := &FTClient{file: file}
	ftClient.Init()

	fmt.Println("Client Started")

	fileSizeStr := strconv.FormatInt(file.size, 10)

	ftClient.tcpClient.Send("FileSize:" + fileSizeStr)

	go tcp.InitTCPServer(":4000", func(data string) {
		if data == "START_SENDING" {
			fmt.Println(data)
			ftClient.sendFile()
			return
		}
		receivedCounts := strings.Split(data, ",")
		log.Println(len(receivedCounts))

		for _, v := range receivedCounts {
			ftClient.tags[v] = true
		}

		tagNotSent := false
		for k, v := range ftClient.tags {

			if k != "" && v != true {
				ftClient.sendFilePart(k)
				tagNotSent = true
			}
		}

		log.Println(tagNotSent)

		if tagNotSent == true {
			ftClient.tcpClient.Send("PASS")
		} else {
			ftClient.file.Close()
			ftClient.tcpClient.Send("EXIT")
			ftClient.tcpClient.Close()
			ftClient.chan_prog <- true
		}

	})

	// ftClient.sendFile()

	<-ftClient.chan_prog

}

func (fileHandle *FileHandle) Init() {
	f, errOpen := os.Open(fileHandle.filePath)
	if errOpen != nil {
		log.Fatalf("Error in opening the file")
	}

	fileHandle.file = f

	fileInfo, err := f.Stat()
	if err != nil {
		log.Fatal("Unable to read file metadata")
	}

	fileHandle.size = fileInfo.Size()
}

func (fileHandle *FileHandle) ReadAt(b []byte, off int64) (n int, err error) {
	return fileHandle.file.ReadAt(b, off)
}

func (fileHandle *FileHandle) Read(b []byte) (n int, err error) {
	return fileHandle.file.Read(b)
}

func (fileHandle *FileHandle) Close() {
	fileHandle.file.Close()
}

func (ftClient *FTClient) Init() {

	ftClient.tags = make(map[string]bool)

	ftClient.udpClient = &udp.UdpClient{}
	ftClient.udpClient.Init(GServer_Address + ":1234")

	ftClient.tcpClient = &tcp.FTTcpClient{}
	ftClient.tcpClient.Init(GServer_Address + ":12341")

	ftClient.chan_prog = make(chan bool)
}

func (ftClient *FTClient) sendFilePart(v string) {
	MAX_LEN := 10 * 1024
	data := make([]byte, MAX_LEN)
	offset, _ := strconv.Atoi(v)
	bytesRead, errRead := ftClient.file.ReadAt(data, int64(offset*MAX_LEN))

	if errRead != nil {
		log.Println("sendFilePart: " + errRead.Error())
	}

	data = data[0:bytesRead]

	counterBytes := intToBytes(int32(offset))

	dataBuffer := append(counterBytes, data...)

	ftClient.udpClient.Send(dataBuffer)
}

func (ftClient *FTClient) sendFile() {
	total := 0
	count := 0
	MAX_LEN := 10 * 1024
	for {
		data := make([]byte, MAX_LEN)
		bytesRead, errRead := ftClient.file.file.Read(data)

		if errRead == io.EOF {
			fmt.Printf("Client: sendFile EOF %v \n", count)
			ftClient.tcpClient.Send("PASS")
			break
		}

		data = data[0:bytesRead]

		counterBytes := intToBytes(int32(count))

		dataBuffer := append(counterBytes, data...)

		ftClient.udpClient.Send(dataBuffer)

		ftClient.tags[strconv.Itoa(count)] = false

		count++

		total += len(dataBuffer)

		// fmt.Println("bytes written to file " + strconv.Itoa(total))
	}
}

func intToBytes(num int32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func bytesToInt(arrByte []byte) uint32 {
	data := binary.BigEndian.Uint32(arrByte)
	return data
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
