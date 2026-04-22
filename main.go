package main 

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	 allBackends = []string{":8081",":8082"}
	 healthyBackends = []string{":8081",":8082"}
   mu sync.RWMutex
	 counter uint64 = 0
)

func main() {
	listenAddr := ":8080"
	go startHealthCheckLoop()

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to bind to port %s : %v" , listenAddr, err)
	}
	defer listener.Close()

	fmt.Printf("Layer 4 load balancer running on TCP port %s\n", listenAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection : %v\n", err)
			continue
		}
		go handleConnection(clientConn)
	}
}

func startHealthCheckLoop() {
	  ticker := time.NewTicker(5 * time.Second)

		for {
			<-ticker.C
			var aliveServers []string 
			for _, backend := range allBackends {
				timeout := 2 * time.Second
				conn,err := net.DialTimeout("tcp", backend, timeout)
				if err != nil {
            fmt.Printf("Health check : %s is down !\n", backend)
				} else{
            conn.Close()
						aliveServers = append(aliveServers, backend)
				}
			}
			mu.Lock()
			healthyBackends = aliveServers
			mu.Unlock()
		}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	clientIP := clientConn.RemoteAddr().String()

	mu.RLock()
	if len(healthyBackends) == 0 {
		mu.RUnlock()
		fmt.Printf("Dropping connection from %s : NO healthy backends available!\n", clientIP)
		clientConn.Write([]byte("503 service unavailable : no backends alive\n"))
		return
	}
	currentCount := atomic.AddUint64(&counter, 1)
	index := currentCount % uint64(len(healthyBackends))
	selectedBackend := healthyBackends[index]

	mu.RUnlock()

	fmt.Printf("routing client %s to backend %s\n", clientIP, selectedBackend)

	backendConn , err := net.Dial("tcp", selectedBackend)
	if err != nil {
		log.Printf("failed to connect to backend %s : %v\n", selectedBackend, err)
		return
	}
	defer backendConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func(){
		defer wg.Done()
		io.Copy(backendConn, clientConn)
	}()

	go func(){
		defer wg.Done()
		io.Copy(clientConn, backendConn)
	}()
	wg.Wait()
}