package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	var requestBuffer []byte
	//requestBuffer := make([]byte, 32)
	//_, err = conn.Read(requestBuffer)
	//if err != nil {
	//	fmt.Println("Cant read connections", err.Error())
	//}

	requestBuffer = make([]byte, 128) // read some data
	conn.Read(requestBuffer)
	firstSpaceIndx := bytes.IndexByte(requestBuffer, ' ')
	requestBuffer = requestBuffer[firstSpaceIndx+1:]
	secondSpaceIndx := bytes.IndexByte(requestBuffer, ' ')
	path := requestBuffer[:secondSpaceIndx]

	okResponse := "HTTP/1.1 200 OK\r\n\r\n"
	notFoundResponse := "HTTP/1.1 404 Not Found\r\n\r\n"
	fmt.Println("path:", string(path))
	if string(path) == "/" {
		sendResponse(okResponse, conn)
	} else if len(path) >= 5 && string(path[:5]) == "/echo" {
		echoRespose := fmt.Sprintf("HTTP/1.1 200 OK \r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %v\r\n"+
			"\r\n"+
			"%s",
			len(path)-6,
			path[6:])
		sendResponse(echoRespose, conn)
	} else {
		sendResponse(notFoundResponse, conn)
	}

}

func sendResponse(response string, conn net.Conn) {
	writeBuffer := []byte(response)
	_, err := conn.Write(writeBuffer)

	if err != nil {
		fmt.Println("Error writing data on connection", err.Error())
	}
	os.Exit(1)
}
