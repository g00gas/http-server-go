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

const (
	HTTP_OK        = "HTTP/1.1 200 OK\r\n"
	HTTP_NOT_FOUND = "HTTP/1.1 404 NOT FOUND\r\n"
	HTTP_CREATED   = "HTTP/1.1 201 CREATED"
)

type HttpRequest struct {
	httpVersion string
	method      string
	path        string
	headers     map[string]string
	body        string
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
	reqObject.body = reqArray[len(reqArray)-1]
	return reqObject
}

func handleConnection(c net.Conn, dir string) {

	buffer := make([]byte, 512)
	_, err := c.Read(buffer)
	if err != nil {
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
	case matchRoute(req.path, `\/files(\/.*)`):
		handleFiles(&req, c, dir)
	default:
		c.Write([]byte(HTTP_NOT_FOUND + "\r\n"))
	}

}

func main() {
	dir := *getArgs()

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
		go func(c net.Conn, dir string) {
			handleConnection(c, dir)
			connection.Close()

		}(connection, dir)
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

func handleFiles(r *HttpRequest, c net.Conn, dir string) {
	pattern := regexp.MustCompile(`/files/(.*)`)
	param := pattern.FindStringSubmatch(r.path)[1]
	path := fmt.Sprintf("%s/%s", dir, param)
	if path == "" {
		c.Write([]byte(HTTP_NOT_FOUND + "\r\n"))
		return
	}
	switch r.method {
	case "GET":
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
	case "POST":
		fmt.Println("Post firing")
		file, err := os.Create(path)
		defer file.Close()
		if err != nil {
			c.Write([]byte(HTTP_NOT_FOUND + "\r\n"))
		}
		content := []byte(r.body)
		fmt.Println(content)
		_, err = file.Write(content)
		if err != nil {
			c.Write([]byte(HTTP_NOT_FOUND + "\r\n"))
		}
		c.Write([]byte(HTTP_CREATED + "\r\n"))
	}
}

func getArgs() *string {
	dir := flag.String("directory", "./", "Directory to get files from")
	flag.Parse()
	return dir
}
