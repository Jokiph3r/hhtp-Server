package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const maxConnections = 10

// MIME types
var mimeTypes = map[string]string{
	".html": "text/html",
	".txt":  "text/plain",
	".gif":  "image/gif",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".css":  "text/css",
}

var connCount = 0
var connLock sync.Mutex

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: http_server <port>")
		return
	}
	port := os.Args[1]

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		connLock.Lock()
		if connCount >= maxConnections {
			connLock.Unlock()
			conn.Close()
			continue
		}
		connCount++
		connLock.Unlock()

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer func() {
		connLock.Lock()
		connCount--
		connLock.Unlock()
	}()

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		httpError(conn, 400, "Bad Request")
		return
	}

	method, path, valid := parseRequestLine(requestLine)
	if !valid {
		httpError(conn, 400, "Bad Request")
		return
	}

	switch method {
	case "GET":
		handleGet(conn, path)
	case "POST":
		handlePost(conn, path, reader)
	default:
		httpError(conn, 501, "Not Implemented")
	}
}

func parseRequestLine(line string) (string, string, bool) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func handleGet(conn net.Conn, path string) {
	filePath := filepath.Join("files", path)
	ext := filepath.Ext(filePath)
	contentType, supported := mimeTypes[ext]
	if !supported {
		httpError(conn, 400, "Bad Request")
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		httpError(conn, 404, "Not Found")
		return
	}
	defer file.Close()

	conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	conn.Write([]byte("Content-Type: " + contentType + "\r\n\r\n"))

	io.Copy(conn, file)
}

func handlePost(conn net.Conn, path string, reader *bufio.Reader) {
	filePath := filepath.Join("files", path)
	file, err := os.Create(filePath)
	if err != nil {
		httpError(conn, 500, "Internal Server Error")
		return
	}
	defer file.Close()

	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		httpError(conn, 500, "Internal Server Error")
		return
	}

	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}

func httpError(conn net.Conn, status int, message string) {
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", status, message)))
}
