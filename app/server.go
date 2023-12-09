package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

func main() {

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

	message := readMessage(conn)
	path := message.getPath()

	okResponse := "HTTP/1.1 200 OK\r\n\r\n"
	notFoundResponse := "HTTP/1.1 404 Not Found\r\n\r\n"

	fmt.Println("path:", string(path))
	if path == "/" {
		sendResponse(okResponse, conn)
	} else if strings.HasPrefix(path, "/echo") {
		echoResponse := fmt.Sprintf("HTTP/1.1 200 OK \r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %v\r\n"+
			"\r\n"+
			"%s",
			len(path)-6,
			path[6:])
		sendResponse(echoResponse, conn)
	} else if strings.HasPrefix(path, "/user-agent") {
		userAgentResponse := fmt.Sprintf("HTTP/1.1 200 OK \r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %v\r\n"+
			"\r\n"+
			"%s",
			len(message.Headers["User-Agent"]),
			message.Headers["User-Agent"])
		sendResponse(userAgentResponse, conn)
	} else {
		sendResponse(notFoundResponse, conn)
	}

}

type Message struct {
	StatusLine string
	Headers    map[string]string
	Body       string
}

func readMessage(conn net.Conn) Message {
	reader := bufio.NewReader(conn)
	message := Message{}
	// read StatusLine
	statusLine, err := reader.ReadString('\n')
	message.StatusLine = statusLine
	if err != nil {
		fmt.Println("Error reading status line", err.Error())
	}

	// read Headers
	message.Headers = make(map[string]string, 8)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading Headers", err.Error())
		}
		if line == "\r\n" { // body and header separator
			break
		}
		key, val := parseLine(line)
		message.Headers[key] = val
	}
	// body not parsed in this stage
	return message
}

func parseLine(line string) (string, string) {
	// Split the line by the first occurrence of ":"
	parts := strings.SplitN(line, ":", 2)
	// Check if the split resulted in a key and value
	if len(parts) != 2 {
		return "", ""
	}
	// Trim whitespaces from key and value
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	return key, value
}

func sendResponse(response string, conn net.Conn) {
	writeBuffer := []byte(response)
	_, err := conn.Write(writeBuffer)

	if err != nil {
		fmt.Println("Error writing data on connection", err.Error())
	}
	os.Exit(1)
}

func (message *Message) getPath() string {
	statusLine := message.StatusLine
	firstSpaceIndx := strings.IndexByte(statusLine, ' ')
	statusLine = statusLine[firstSpaceIndx+1:]
	secondSpaceIndx := strings.IndexByte(statusLine, ' ')
	path := statusLine[:secondSpaceIndx]
	return path
}
