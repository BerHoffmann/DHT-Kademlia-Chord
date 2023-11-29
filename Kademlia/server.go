package main

import (
    "fmt"
    "net"
	//"encoding/json"
	//"log"
	//"time"
)

var(
	hashSize = 6 // 2^6 nodes
    bucketSize = 2
    distanceToStore = 4
	newNode Node = Node{}
	myNode Node = Node{}
    distanceTable = [6][]Node{}
)

func main() {
    listener, err := net.Listen("tcp", ":8123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server is listening on port 8123")
	
	//go HandleUserInput()

    for {
        // Accept incoming connections
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error:", err)
            continue
        }

		go HandleMessage(conn)
	}
    
}
