package main

import (
    "fmt"
    "net"
	//"encoding/json"
	//"log"
	//"time"
)

func (n *NodeConfig) createNewRing() {
	n.Precessor.HashPosition = -1
	n.Sucessor.HashPosition = -1
}

func main() {

    N := new(NodeConfig) // ponteiro, nao struct de fato
    N.createNewRing()
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

		go N.HandleMessage(conn)
	}
}



//conn.RemoteAddr() and LocalAddr()