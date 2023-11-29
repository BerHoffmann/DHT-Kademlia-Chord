package main

import (
    "fmt"
    "net"
    //"time"
)


func main() {
    // Listen for incoming connections
    // time.Sleep(2 * time.Second)
    listener, err := net.Listen("tcp", ":8123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server is listening on port 8123")

    for {
        // Accept incoming connections
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error:", err)
            continue
        }

        // Handle client connection in a goroutine
        go handleClient(conn)
    }
}


func handleClient(conn net.Conn) {
    defer conn.Close()

    buf := make([]byte, 1024)
    
    for {
        // se conexao fechar no client, da erro de EOF
        n, err := conn.Read(buf)
        if err != nil {
            fmt.Println("Error:", err)
           return
        }
        
    remoteAddr := conn.RemoteAddr()
     
    
    fmt.Printf("Received: %s from IP %s\n", buf[:n], remoteAddr)
    //time.Sleep(2 * time.Second)
    }
}






