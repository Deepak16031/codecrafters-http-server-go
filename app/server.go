package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}

}

const (
	Ok_RESPONSE             = "HTTP/1.1 200 OK"
	CREATED_RESPONSE        = "HTTP/1.1 201 CREATED"
	NOT_FOUND_RESPONSE      = "HTTP/1.1 404 Not Found"
	CONTENT_LENGTH          = "Content-Length"
	CONTENT_TYPE_TEXT_PLAIN = "Content-Type: text/plain"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	message := readMessage(conn)
	path := message.getPath()
	okResponse := "HTTP/1.1 200 OK\r\n\r\n"
	notFoundResponse := "HTTP/1.1 404 Not Found\r\n\r\n"

	requestType := message.getRequestType()
	if requestType == "POST" {
		var directoryToServe string
		if len(os.Args) == 3 && os.Args[1] == "--directory" {
			directoryToServe = os.Args[2]
		}
		fileName := path[6:]
		fmt.Println("directoryToServe:", directoryToServe, "fileName:", fileName)
		filePath := directoryToServe + fileName
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			fmt.Println("Error opening file:", filePath, "Error:", err.Error())
		}
		file.Write([]byte(message.Body))
		createdResponse := fmt.Sprintf("%s\r\n\r\n", CREATED_RESPONSE)
		sendResponse(createdResponse, conn)
		return
	} else if requestType != "GET" {
		fmt.Println("Not Supported Request Type")
		return
	}
	if path == "/" {
		sendResponse(okResponse, conn)
	} else if strings.HasPrefix(path, "/echo") {
		echoResponse := fmt.Sprintf("HTTP/1.1 200 OK \r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %v\r\n"+
			"\r\n"+
			"%s",
			len(path)-6,
			path[6:]) // path[5] is /
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
	} else if strings.HasPrefix(path, "/files") {
		var directoryToServe string
		if len(os.Args) == 3 && os.Args[1] == "--directory" {
			directoryToServe = os.Args[2]
		}
		fileName := path[6:]
		fmt.Println("directoryToServe:", directoryToServe, "fileName:", fileName)
		filePath := directoryToServe + fileName[1:]
		fmt.Println("FilePath:", filePath)
		sendFile(conn, filePath)
	} else {
		sendResponse(notFoundResponse, conn)
	}
}

func sendFile(conn net.Conn, filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading filePath:", filePath)
		sendResponse(fmt.Sprintf("%s\r\n\r\n", NOT_FOUND_RESPONSE), conn)
		return
	}
	response := fmt.Sprintf("%s\r\n", Ok_RESPONSE) +
		"Content-Type: application/octet-stream\r\n" +
		fmt.Sprintf("Content-Length: %v\r\n", len(data)) +
		"\r\n" + string(data)
	sendResponse(response, conn)
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
	if val, ok := message.Headers[CONTENT_LENGTH]; ok {
		size, err := strconv.Atoi(val)
		if err != nil {
			fmt.Println("Cant convert string to int", err.Error())
		}

		bodyBuffer := make([]byte, size)
		io.ReadFull(reader, bodyBuffer)
		message.Body = string(bodyBuffer)
	}
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
}

func (message *Message) getPath() string {
	statusLine := message.StatusLine
	firstSpaceIndx := strings.IndexByte(statusLine, ' ')
	statusLine = statusLine[firstSpaceIndx+1:]
	secondSpaceIndx := strings.IndexByte(statusLine, ' ')
	path := statusLine[:secondSpaceIndx]
	return path
}
func (message *Message) getRequestType() string {
	statusLine := message.StatusLine
	firstSpaceIndx := strings.IndexByte(statusLine, ' ')
	requestType := message.StatusLine[:firstSpaceIndx]
	return requestType
}
