package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

var backends = []string{":8081", ":8082"} // The addresses of our dummy backend servers
var counter uint64 = 0                    // An atomic counter to track connections for Round Robin

func main() {
	// 1. Define the port our Load Balancer will listen on.
	listenAddr := ":8080"

	// 2. Start listening for pure, raw TCP connections.
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		// Note: log.Fatalf automatically calls os.Exit(1) for you!
		log.Fatalf("Failed to bind to port %s: %v", listenAddr, err)
	}
	defer listener.Close()

	fmt.Printf("Layer 4 Load Balancer is running on TCP port %s\n", listenAddr)

	// 3. The Infinite Listener Loop
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		// 4. Concurrency Magic: Spawn a goroutine for every single user.
		go handleConnection(clientConn)
	}
}


func handleConnection(clientConn net.Conn) {
	
	defer clientConn.Close()

	clientIP := clientConn.RemoteAddr().String()
	fmt.Printf("New connection received from client: %s\n", clientIP)

	// exact same millisecond, sync/atomic ensures this is completely thread-safe.
	currentCount := atomic.AddUint64(&counter, 1)
	
	// Modulo math ensures the index always wraps around (0, 1, 0, 1...)
	index := currentCount % uint64(len(backends))
	selectedBackend := backends[index]
	
	fmt.Printf("Routing client %s to backend %s\n", clientIP, selectedBackend)

	// The load balancer now acts as a client, reaching out to the chosen backend server.
	backendConn, err := net.Dial("tcp", selectedBackend)
	if err != nil {
		log.Printf("Failed to connect to backend %s: %v\n", selectedBackend, err)
		return // If the backend is down, abort handling this client.
	}
	
	defer backendConn.Close()

	// using a WaitGroup to prevent the function from exiting (and closing our connections) until both data streams finish their work.
	var wg sync.WaitGroup
	wg.Add(2)

	// Stream 1: Client -> Backend
	go func() {
		defer wg.Done() // Signal that this stream is finished when io.Copy unblocks
		io.Copy(backendConn, clientConn)
	}()

	// Stream 2: Backend -> Client
	go func() {
		defer wg.Done() // Signal that this stream is finished when io.Copy unblocks
		io.Copy(clientConn, backendConn)
	}()

	// Wait here, blocking this specific goroutine, until both wg.Done() calls are made.
	wg.Wait()
	
	fmt.Printf("Connection closed for client: %s\n", clientIP)
}