package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "sync"
    "time"

    pb "kvstore/proto"
    "google.golang.org/grpc"
)

var peerAddrs = []string{
    "localhost:5051",
    "localhost:5052",
    "localhost:5053",
}

var myAddr string // set from command line arg

type kvServer struct {
    pb.UnimplementedKeyValueStoreServer
    mu    sync.Mutex
    store map[string]string
}

func (s *kvServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
    log.Printf("[PUT] Key=%s Value=%s IsReplica=%v", req.Key, req.Value, req.IsReplica)

    s.mu.Lock()
    s.store[req.Key] = req.Value
    s.mu.Unlock()

    if !req.IsReplica {
        for _, addr := range peerAddrs {
            if addr == myAddr {
                continue
            }
            go replicatePut(addr, req.Key, req.Value)
        }
    }
    return &pb.PutResponse{Message: "OK"}, nil
}

func replicatePut(addr, key, value string) {
    log.Printf("[REPLICATE PUT] to %s Key=%s Value=%s", addr, key, value)
    conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
    if err != nil {
        log.Printf("Failed to connect to peer %s: %v", addr, err)
        return
    }
    defer conn.Close()

    client := pb.NewKeyValueStoreClient(conn)
    _, err = client.Put(context.Background(), &pb.PutRequest{
        Key:       key,
        Value:     value,
        IsReplica: true,
    })
    if err != nil {
        log.Printf("Failed to replicate put to %s: %v", addr, err)
    }
}

func (s *kvServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
    s.mu.Lock()
    val, ok := s.store[req.Key]
    s.mu.Unlock()
    log.Printf("[GET] Key=%s Found=%v", req.Key, ok)
    return &pb.GetResponse{Value: val, Found: ok}, nil
}

func (s *kvServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
    log.Printf("[DELETE] Key=%s IsReplica=%v", req.Key, req.IsReplica)

    s.mu.Lock()
    delete(s.store, req.Key)
    s.mu.Unlock()

    if !req.IsReplica {
        for _, addr := range peerAddrs {
            if addr == myAddr {
                continue
            }
            go replicateDelete(addr, req.Key)
        }
    }
    return &pb.DeleteResponse{Message: "Deleted"}, nil
}

func replicateDelete(addr, key string) {
    log.Printf("[REPLICATE DELETE] to %s Key=%s", addr, key)
    conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
    if err != nil {
        log.Printf("Failed to connect to peer %s: %v", addr, err)
        return
    }
    defer conn.Close()

    client := pb.NewKeyValueStoreClient(conn)
    _, err = client.Delete(context.Background(), &pb.DeleteRequest{
        Key:       key,
        IsReplica: true,
    })
    if err != nil {
        log.Printf("Failed to replicate delete to %s: %v", addr, err)
    }
}

func (s *kvServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // Copy to avoid race
    copyStore := make(map[string]string, len(s.store))
    for k, v := range s.store {
        copyStore[k] = v
    }
    log.Printf("[LIST] Returning %d keys", len(copyStore))
    return &pb.ListResponse{Store: copyStore}, nil
}

// syncFromPeer tries to fetch store data from one peer and merges it in this node
func (s *kvServer) syncFromPeer(peer string) error {
    log.Printf("[SYNC] Syncing from peer %s", peer)
    conn, err := grpc.Dial(peer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
    if err != nil {
        return fmt.Errorf("failed to connect to peer %s: %w", peer, err)
    }
    defer conn.Close()

    client := pb.NewKeyValueStoreClient(conn)
    resp, err := client.List(context.Background(), &pb.ListRequest{})
    if err != nil {
        return fmt.Errorf("failed to list from peer %s: %w", peer, err)
    }

    s.mu.Lock()
    defer s.mu.Unlock()
    for k, v := range resp.Store {
        if _, exists := s.store[k]; !exists {
            log.Printf("[SYNC] Adding missing key %s from peer", k)
            s.store[k] = v
        }
    }
    return nil
}

func main() {
    port := "5051"
    if len(os.Args) > 1 {
        port = os.Args[1]
    }
    myAddr = "localhost:" + port

    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        log.Fatalf("Failed to listen on %s: %v", port, err)
    }

    server := &kvServer{
        store: make(map[string]string),
    }

    // Fault simulation: On startup, try to sync missing keys from first available peer (excluding self)
    go func() {
        time.Sleep(3 * time.Second) // wait a bit for other nodes to start
        for _, peer := range peerAddrs {
            if peer == myAddr {
                continue
            }
            if err := server.syncFromPeer(peer); err == nil {
                log.Printf("[SYNC] Successfully synced from peer %s", peer)
                break
            } else {
                log.Printf("[SYNC] Failed to sync from peer %s: %v", peer, err)
            }
        }
    }()

    grpcServer := grpc.NewServer()
    pb.RegisterKeyValueStoreServer(grpcServer, server)

    log.Printf("Server listening on port :%s", port)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
