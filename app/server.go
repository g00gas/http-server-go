package main

import (
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

type HttpRequest struct {
	httpVersion string
	method      string
	path        string
	headers     []string
}

func returnHttpRequest(reqBuffer []byte) HttpRequest {
	reqObject := HttpRequest{}
	reqArray := strings.Split(string(reqBuffer), "\r\n")
	startLine := strings.Split(reqArray[0], " ")
	reqObject.method = startLine[0]
	reqObject.path = startLine[1]
	reqObject.httpVersion = startLine[2]
	for _, header := range reqArray[3:] {
		if header != "" {
			reqObject.headers = append(reqObject.headers, header)
		}
	}
	return reqObject
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	connection, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	buffer := make([]byte, 512)
	_, err = connection.Read(buffer)
	if err != nil {
		fmt.Println("error reading connection ", err.Error())
		os.Exit(1)
	}
	req := returnHttpRequest(buffer)
	switch req.path {
	case "/":
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	default:
		connection.Write([]byte("HTTP/1.1 404 NOT FOUND\r\n\r\n"))
	}

	connection.Close()

}
