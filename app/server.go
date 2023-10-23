package main

import (
	"fmt"
	"regexp"
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

const (
	HTTP_OK        = "HTTP/1.1 200 OK\r\n"
	HTTP_NOT_FOUND = "HTTP/1.1 404 NOT FOUND\r\n"
)

func main() {
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
	switch {
	case matchRoute(req.path, `/`):
		connection.Write([]byte(HTTP_OK))
	case matchRoute(req.path, `\/echo(\/.*)`):
		handleEcho(&req, connection)
	default:
		connection.Write([]byte(HTTP_NOT_FOUND))
	}

	connection.Close()

}

func matchRoute(requestPath, pattern string) bool {
	regex := regexp.MustCompile("^" + pattern + "$")
	match := regex.FindStringSubmatch(requestPath)
	if match != nil {
		return true
	} else {
		return false
	}
}

func handleEcho(r *HttpRequest, c net.Conn) {
	pattern := regexp.MustCompile(`/echo/(.*)`)
	param := pattern.Find([]byte(r.path))
	headers := strings.Join([]string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(param))}, "\r\n")
	res := HTTP_OK + headers + fmt.Sprintf("\r\n%s", param)
	c.Write([]byte(res))
}
