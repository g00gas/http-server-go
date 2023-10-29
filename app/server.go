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
	headers     map[string]string
}

func returnHttpRequest(reqBuffer []byte) HttpRequest {
	reqObject := HttpRequest{}
	reqArray := strings.Split(string(reqBuffer), "\r\n")
	startLine := strings.Split(reqArray[0], " ")
	reqObject.method = startLine[0]
	reqObject.path = startLine[1]
	reqObject.httpVersion = startLine[2]
	reqObject.headers = make(map[string]string)
	for _, header := range reqArray[1:] {
		if header != "" {
			v := strings.Split(header, ": ")
			if len(v) > 1 {
				reqObject.headers[v[0]] = v[1]
			}
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
		connection.Write([]byte(HTTP_OK + "\r\n"))
	case matchRoute(req.path, `\/echo(\/.*)`):
		handleEcho(&req, connection)
	case matchRoute(req.path, `\/user-agent`):
		handleUserAgent(&req, connection)
	default:
		connection.Write([]byte(HTTP_NOT_FOUND + "\r\n"))
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
	param := pattern.FindStringSubmatch(r.path)
	if len(param) > 1 {
		headers := strings.Join([]string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(param[1]))}, "\r\n")
		res := HTTP_OK + headers + fmt.Sprintf("\r\n\r\n%s", param[1])
		c.Write([]byte(res))
	}
}

func handleUserAgent(r *HttpRequest, c net.Conn) {
	agent := r.headers["User-Agent"]
	if len(agent) > 1 {
		headers := strings.Join([]string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(agent))}, "\r\n")
		res := HTTP_OK + headers + fmt.Sprintf("\r\n\r\n%s", agent)
		c.Write([]byte(res))
	}
}
