# http-Server
A http server implemneted in Go language 


For Server on EC2 go-http-server : 

Install Go:  sudo apt install golang -y

Verify Go installation: go version

Deploy the HTTP Server

    Create the Project:
        SSH into the EC2 instance and create a directory for the project: mkdir http_server && cd http_server

    Create the main Go file:  nano server.go

    Copy the Go server code from the implementation above into the server.go file. Save and exit.

    Set Up the Files Directory: Create a directory to store files for POST/GET operations: mkdir files

Compile the Server:

    Build the server binary: go build -o http_server server.go

Run the Server:

    Run the server on a specified port (e.g., 8080): ./http_server 8080

The server will now listen on port 8080.

For Proxy : 

    Commands to Compile and Run

    Compile the Proxy Server: go build -o proxy proxy.go

    Run the Proxy Server:  ./proxy <port>

    Replace <port> with the port number you want the proxy to listen on, e.g., 8080.


Test from Client:

curl -X GET http://<server_ip>:8080/index.html -x <proxy_ip>:8081



