package main

import (
	"flag"
	"fmt"
	"io"
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

func handleConnection(c net.Conn) {

	buffer := make([]byte, 512)
	_, err := c.Read(buffer)
	if err != nil {
		fmt.Println("error reading c ", err.Error())
		os.Exit(1)
	}
	req := returnHttpRequest(buffer)
	switch {
	case matchRoute(req.path, `/`):
		c.Write([]byte(HTTP_OK + "\r\n"))
	case matchRoute(req.path, `\/echo(\/.*)`):
		handleEcho(&req, c)
	case matchRoute(req.path, `\/user-agent`):
		handleUserAgent(&req, c)
	case matchRoute(req.path, `\/file(\/.*)`):
		handleFiles(&req, c)
	default:
		c.Write([]byte(HTTP_NOT_FOUND + "\r\n"))
	}

}

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go func(c net.Conn) {
			handleConnection(c)
			connection.Close()

		}(connection)
	}

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

func handleFiles(r *HttpRequest, c net.Conn) {
	dir := *getArgs()
	pattern := regexp.MustCompile(`/file/(.*)`)
	param := pattern.FindStringSubmatch(r.path)[1]
	fmt.Printf(fmt.Sprintf("%s/%s", dir, param))
	if len(param) > 1 {
		path := fmt.Sprintf("%s/%s", dir, param)
		file, err := os.Open(path)
		if err != nil {
			c.Write([]byte(HTTP_NOT_FOUND + "\r\n"))
		}
		defer file.Close()
		buffer, err := io.ReadAll(file)
		if err != nil {
			fmt.Println("cannot read file")
		}
		content := string(buffer)
		headers := strings.Join([]string{"Content-Type: application/octet-stream", fmt.Sprintf("Content-Length: %d", len(content))}, "\r\n")
		res := HTTP_OK + headers + fmt.Sprintf("\r\n\r\n%s", content)
		c.Write([]byte(res))
	}
}

func getArgs() *string {
	dir := flag.String("directory", "./", "Directory to get files from")
	flag.Parse()
	return dir
}
