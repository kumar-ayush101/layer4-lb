package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	conn , err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Failed to connect to load balancer : %v" , err)
	}
	defer conn.Close()

	fmt.Println("Connected to Load Balancer!")

	fmt.Fprintf(conn, "Hello from the test client!")
  
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read response : %v" , err)
	}

	fmt.Printf("Received response : %s", response)

}