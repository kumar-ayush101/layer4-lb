package main 

import (
	"encoding/json" 
	"fmt"
	"io"
	"log"
	"net"
	"os"           
	"sync"
	"sync/atomic"
	"time"
)


type config struct {
	ListenPort string   `json:"listen_port"`
	Backends   []string `json:"backends"`
}

var (
	allBackends     []string
	healthyBackends []string
	mu              sync.RWMutex
	counter         uint64 = 0
)

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config.json: %v\n", err)
	}

	
	allBackends = cfg.Backends
	healthyBackends = cfg.Backends 
	listenAddr := cfg.ListenPort

	fmt.Printf("Loaded %d backends from config.json\n", len(allBackends))

	
	go startHealthCheckLoop()

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to bind to port %s : %v\n" , listenAddr, err)
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


func loadConfig(filename string) (*config, error) {
	fileBytes , err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg config 
	err = json.Unmarshal(fileBytes, &cfg)
	if err != nil {
		return nil , err
	}
	return &cfg, nil
}