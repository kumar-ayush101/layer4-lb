# Go Layer 4 TCP Load Balancer



A highly concurrent, self-healing Layer 4 (Transport Layer) Load Balancer written entirely in Go. 

Unlike Layer 7 proxies that parse HTTP headers, this load balancer operates directly at the TCP level. It accepts raw network sockets, routes traffic using a Round-Robin algorithm, and efficiently pumps full-duplex byte streams between clients and backend servers.

## ✨ Features

* **Pure Layer 4 Proxying:** Agnostic to application-layer protocols (works with HTTP, SSH, raw TCP, custom game protocols, etc.).
* **High Concurrency:** Utilizes Go's lightweight Goroutines to handle thousands of simultaneous client connections without blocking.
* **Full-Duplex Streaming:** Implements bi-directional byte pumping using `io.Copy` for zero-allocation, highly efficient data transfer.
* **Active Health Checks:** Runs a background monitoring loop to ping backend servers. Automatically removes dead servers from the routing pool and re-adds them when they recover.
* **Thread-Safe State:** Uses `sync.RWMutex` to safely allow background health checks to modify the backend pool while concurrent foreground traffic reads from it.
* **Lockless Round-Robin:** Uses `sync/atomic` counters to route traffic across backends without CPU-blocking mutex locks.
* **Dynamic Configuration:** Reads listening ports and backend addresses from a `config.json` file on startup.

## 📂 Project Structure

* `/main.go` - The core Load Balancer engine.
* `/config.json` - Configuration file for dynamic port binding and backend discovery.
* `/backend/backend.go` - A lightweight dummy TCP server used to simulate application backends.
* `/client/test_client.go` - A simulated TCP client to test routing and responses.

## 🚀 How to Run Locally

To see the load balancer, health checks, and Round-Robin routing in action, you will need 4 terminal windows.

### 1. Configure the Load Balancer
Ensure your `config.json` looks like this in the root directory:

```json
{
  "listen_port": ":8080",
  "backends": [
    ":8081",
    ":8082"
  ]
}
```

### 2. Start the Backend Servers
Open two terminal windows to act as your application servers.

**Terminal 1:**
```bash
cd backend
go run backend.go :8081
```

**Terminal 2:**
```bash
cd backend
go run backend.go :8082
```

### 3. Start the Load Balancer
Open a third terminal in the root directory to start the proxy.

**Terminal 3:**
```bash
go run main.go
```
*You will see the load balancer start up and begin actively health-checking the backends.*

### 4. Send Client Traffic
Open a fourth terminal to simulate users connecting to your system. Run this command multiple times.

**Terminal 4:**
```bash
cd client
go run test_client.go
```

**Expected Output:** You will see the responses alternate perfectly between `:8081` and `:8082`!

## 🩺 Testing the Self-Healing (Health Checks)

1. While the load balancer is running, go to **Terminal 1** (Backend `:8081`) and press `Ctrl+C` to crash the server.
2. Watch the Load Balancer terminal. Within 5 seconds, it will detect the failure and print: `🩺 Health Check: :8081 is down!`
3. Run the client (`test_client.go`) multiple times. Notice that **100% of traffic now safely routes to `:8082`**. No connections are dropped!
4. Restart Backend `:8081`. The load balancer will automatically detect its recovery and resume Round-Robin routing.

## 🛠️ Built With
* **Go (Golang)** - Standard library only (`net`, `io`, `sync`, `sync/atomic`, `encoding/json`). No external dependencies.