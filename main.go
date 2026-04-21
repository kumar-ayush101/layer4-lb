package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// 1. Define the port our Load Balancer will listen on.
	// This is the "Public IP/Port" that all clients will connect to.
	listenAddr := ":8080"

	// 2. Start listening for pure, raw TCP connections.
	// We are NOT using net/http. We are operating at Layer 4.
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to bind to port %s: %v", listenAddr, err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("🚀 Layer 4 Load Balancer is running on TCP port %s\n", listenAddr)

	// 3. The Infinite Listener Loop
	for {
		// Accept() blocks until a new client tries to connect.
		// When they do, it gives us a raw network socket (clientConn).
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue // If one user fails, keep listening for the next one
		}

		// 4. The Concurrency Magic!
		// For every single user, we spawn a new Goroutine. 
		// This allows 100,000 users to connect simultaneously without blocking this loop.
		go handleConnection(clientConn)
	}
}

// handleConnection is where the routing and byte-pumping will happen.
func handleConnection(clientConn net.Conn) {
	// We make sure to close the client connection when we are completely done.
	defer clientConn.Close()

	clientIP := clientConn.RemoteAddr().String()
	fmt.Printf("✅ New connection received from client: %s\n", clientIP)

	// TODO: Phase 2 - Use an algorithm to pick a backend server.
	// TODO: Phase 3 - Open a TCP connection to that backend.
	// TODO: Phase 4 - Stream the raw bytes back and forth.
	
	// For now, let's just send a raw byte response to prove it works and then close it.
	clientConn.Write([]byte("Hello from your Layer 4 Load Balancer!\n"))
}