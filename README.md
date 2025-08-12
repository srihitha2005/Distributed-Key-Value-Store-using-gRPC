# Distributed gRPC Key-Value Store with Replication & Fault Tolerance

This is a simple distributed Key-Value Store implemented in Go using gRPC.  
It supports replication across multiple servers, fault tolerance simulation, and detailed logging.

---

## Features

- Basic CRUD operations: `put`, `get`, `delete`, `list` via gRPC  
- Data replication to peer servers for eventual consistency  
- Fault tolerance: simulate node failure and recovery  
- Thread-safe in-memory store with mutex locks  
- Logging of requests and replication events  

---

## Prerequisites

- Go 1.18+ ([https://golang.org/dl/](https://golang.org/dl/))  
- Protocol Buffers compiler (`protoc`) ([https://grpc.io/docs/protoc-installation/](https://grpc.io/docs/protoc-installation/))  
- Go protobuf plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure $GOPATH/bin is added to your system $PATH.

## Setup & Run
### Clone repository

```bash
git clone https://github.com/yourusername/kvstore.git
```

### Generate Go protobuf code
Make sure you have protoc, protoc-gen-go, and protoc-gen-go-grpc installed and in your PATH.

```bash
protoc --go_out=. --go-grpc_out=. proto/kvstore.proto
```

### Run multiple server instances
Open multiple terminals and run the server on different ports:

```bash
go run server/main.go 5051
```

```bash
go run server/main.go 5052
```

```bash
go run server/main.go 5053
```

### Run CLient
Connect client to any running server port:

```bash
go run client/main.go 5051
```

## Usage
In client CLI, use these commands:
```bash
put <key> <value>    - Store a key-value pair
get <key>            - Retrieve value by key
delete <key>         - Remove key-value pair
list                 - List all keys and values on that server
help                 - Show command help
exit                 - Exit client
```
