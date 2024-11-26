package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const maxConnections = 10

var connCount = 0
var connLock sync.Mutex

var supportedFileTypes = map[string]bool{
	".html": true,
	".txt":  true,
	".gif":  true,
	".jpeg": true,
	".jpg":  true,
	".css":  true,
}



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
	defer func() {
		conn.Close() // Ensure the connection is always closed
		connLock.Lock()
		connCount--
		connLock.Unlock()
	}()

	reader := bufio.NewReader(conn)

	// Read the HTTP request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		httpError(conn, 400, "Bad Request")
		return
	}

	// Parse the HTTP request line
	method, path, valid := parseRequestLine(requestLine)
	if !valid {
		httpError(conn, 400, "Bad Request")
		return
	}

	// Parse headers for POST requests
	var contentLength int = -1
	if method == "POST" {
		contentLength = parseHeaders(reader)
		if contentLength < 0 {
			httpError(conn, 411, "Length Required")
			return
		}
	}
	// Route the request
	switch method {
	case "GET":
		handleGet(conn, path)
	case "POST":
		handlePost(conn, path, reader, contentLength)
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

func parseHeaders(reader *bufio.Reader) int {
	var contentLength int = -1

	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if strings.ToLower(key) == "content-length" {
				contentLength, _ = strconv.Atoi(value)
			}
		}
	}
	return contentLength
}

func handleGet(conn net.Conn, path string) {
	filePath := filepath.Join("files", path)
	ext := filepath.Ext(filePath)

	// Validate file type
	if !supportedFileTypes[ext] {
		httpError(conn, 400, "Bad Request: Unsupported file type")
		return
	}

	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		httpError(conn, 404, "Not Found")
		return
	}
	defer file.Close()

	conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	conn.Write([]byte("Content-Type: text/plain\r\n"))
	conn.Write([]byte("Connection: close\r\n\r\n"))
	io.Copy(conn, file)

	conn.Close() // Immediately close the connection after response
	fmt.Println("GET request handled. Connection closed.")
}

func handlePost(conn net.Conn, path string, reader *bufio.Reader, contentLength int) {
	filePath := filepath.Join("files", path)

	// Create or overwrite the file
	file, err := os.Create(filePath)
	if err != nil {
		httpError(conn, 500, "Internal Server Error")
		return
	}
	defer file.Close()

	// Read the exact number of bytes from the body
	_, err = io.CopyN(file, reader, int64(contentLength))
	if err != nil {
		httpError(conn, 500, "Internal Server Error")
		return
	}

	// Send the response
	conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	conn.Write([]byte("Content-Type: text/plain\r\n"))
	conn.Write([]byte("Connection: close\r\n\r\n"))

	conn.Close() // Forcefully close the connection
	fmt.Println("POST request handled. Connection closed.")
}

func httpError(conn net.Conn, status int, message string) {
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, message)))
	conn.Write([]byte("Connection: close\r\n\r\n"))
	conn.Close() // Ensure connection is closed even in case of errors
	fmt.Printf("Error %d: %s. Connection closed.\n", status, message)
}
