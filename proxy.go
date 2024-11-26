package main


import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)

const maxConnections = 10

var connCount = 0
var connLock sync.Mutex

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: proxy <port>")
		return
	}
	port := os.Args[1]

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting proxy:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Proxy started on port", port)

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

		go handleProxyConnection(conn)
	}
}

func handleProxyConnection(clientConn net.Conn) {
	defer clientConn.Close()
	defer func() {
		connLock.Lock()
		connCount--
		connLock.Unlock()
	}()

	reader := bufio.NewReader(clientConn)
	request, err := http.ReadRequest(reader)
	if err != nil {
		httpError(clientConn, 400, "Bad Request")
		return
	}

	if request.Method != "GET" {
		httpError(clientConn, 501, "Not Implemented")
		return
	}

	// Forward the request to the target server
	targetConn, err := net.Dial("tcp", request.Host)
	if err != nil {
		httpError(clientConn, 502, "Bad Gateway")
		return
	}
	defer targetConn.Close()

	// Send the request to the origin server
	err = request.Write(targetConn)
	if err != nil {
		httpError(clientConn, 502, "Bad Gateway")
		return
	}

	// Relay the response from the origin server back to the client
	targetReader := bufio.NewReader(targetConn)
	response, err := http.ReadResponse(targetReader, request)
	if err != nil {
		httpError(clientConn, 502, "Bad Gateway")
		return
	}

	err = response.Write(clientConn)
	if err != nil {
		fmt.Println("Error writing response to client:", err)
	}
}

func httpError(conn net.Conn, status int, message string) {
	fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n", status, message)
	fmt.Fprintf(conn, "Connection: close\r\n\r\n")
}
