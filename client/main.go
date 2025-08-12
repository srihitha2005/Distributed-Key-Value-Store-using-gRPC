package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    pb "kvstore/proto"
    "google.golang.org/grpc"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run client/main.go <server_port>")
        return
    }
    serverAddr := "localhost:" + os.Args[1]

    conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Fatalf("Could not connect to server: %v", err)
    }
    defer conn.Close()

    c := pb.NewKeyValueStoreClient(conn)

    fmt.Println("Connected to gRPC Key-Value Store at", serverAddr)
    fmt.Println("Type 'help' for commands.")

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print(">")
        if !scanner.Scan() {
            break
        }
        input := strings.TrimSpace(scanner.Text())
        parts := strings.SplitN(input, " ", 3)

        if len(parts) == 0 || parts[0] == "" {
            continue
        }

        cmd := strings.ToLower(parts[0])
        ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()

        switch cmd {
        case "put":
            if len(parts) < 3 {
                fmt.Println("Usage: put <key> <value>")
                continue
            }
            _, err := c.Put(ctx, &pb.PutRequest{Key: parts[1], Value: parts[2]})
            if err != nil {
                log.Printf("Error on put: %v\n", err)
            } else {
                fmt.Println("Stored successfully")
            }

        case "get":
            if len(parts) < 2 {
                fmt.Println("Usage: get <key>")
                continue
            }
            res, err := c.Get(ctx, &pb.GetRequest{Key: parts[1]})
            if err != nil {
                log.Printf("Error on get: %v\n", err)
            } else if !res.Found {
                fmt.Println("Key not found")
            } else {
                fmt.Printf("%s = %s\n", parts[1], res.Value)
            }

        case "delete":
            if len(parts) < 2 {
                fmt.Println("Usage: delete <key>")
                continue
            }
            _, err := c.Delete(ctx, &pb.DeleteRequest{Key: parts[1]})
            if err != nil {
                log.Printf("Error on delete: %v\n", err)
            } else {
                fmt.Println("Deleted successfully")
            }

        case "list":
            res, err := c.List(ctx, &pb.ListRequest{})
            if err != nil {
                log.Printf("Error on list: %v\n", err)
            } else {
                if len(res.Store) == 0 {
                    fmt.Println("(empty)")
                } else {
                    fmt.Println("Current store contents:")
                    for k, v := range res.Store {
                        fmt.Printf("%s = %s\n", k, v)
                    }
                }
            }

        case "help":
            fmt.Println("Commands:")
            fmt.Println(" put <key> <value>  - Add or update a key")
            fmt.Println(" get <key>          - Retrieve a key")
            fmt.Println(" delete <key>       - Remove a key")
            fmt.Println(" list               - Show all keys")
            fmt.Println(" exit               - Quit the program")

        case "exit":
            fmt.Println("Bye!")
            return

        default:
            fmt.Println("Unknown command. Type 'help' for commands.")
        }
    }
}
