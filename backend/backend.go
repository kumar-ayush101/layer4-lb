package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Require the binding port to be passed dynamically via CLI.
	if len(os.Args) < 2 {
		log.Fatal("Please provide a port! Example: go run backend.go :8081")
	}
	
	port := os.Args[1]

	// Bind the server to the specified TCP port for raw layer 4 traffic.
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to start backend on %s: %v", port, err)
	}
	
	// Ensure the socket is released back to the OS upon termination.
	defer listener.Close()

	fmt.Printf(" Backend Server running on port %s\n", port)

	// Main event loop: continuously process incoming client connections.
	for {
		// Block until a new TCP handshake is completed.
		conn, err := listener.Accept()
		if err != nil {
			// Log the error but keep the listener alive for future traffic.
			log.Printf("Connection accept error: %v", err)
			continue
		}

		message := fmt.Sprintf("Hello from the backend server running on port %s!\n", port)
		
		// Write the payload as bytes and immediately terminate the connection 
		// (simulate a quick, stateless response).
		conn.Write([]byte(message))
		conn.Close()
	}
}