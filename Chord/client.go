package main

import (
    "fmt"
    "net"
    //"time"
    //"os"
    //"encoding/json"
    //"strings"
    //"strconv"
)

//o cliente, uma vez que entrou no CHORD, vai ter que fazer o mesmo papel que o server, ou seja, 
//ficar ouvindo na porta 8123, como em net.Listen("tcp", ":8123")

func main() {

    N := new(NodeConfig)
    listener, err := net.Listen("tcp", ":8123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer listener.Close()

    go N.HandleUserInput()

    // i, _ := strconv.Atoi(os.Args[2])
    // time.Sleep(time.Duration(i) * time.Second)

    // if len(os.Args) == 5 { //usage: program ip filename filecontent   
    //     go N.sendMessage(os.Args[3], os.Args[4])
    // }

    // if len(os.Args) == 6 { //usage: program ip filename filecontent filetosearch
    //     go N.sendMessage(os.Args[3], os.Args[4])
    //     go N.searchMessage(os.Args[5])
    // }

    // -------------------------------------------------------

    fmt.Println("Server is listening on port 8123")
	
	//go HandleUserInput()

    for {
        // Accept incoming connections
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error:", err)
            continue
        }

       // fmt.Println("Going to handle msg")
		go N.HandleMessage(conn)
	}
    // -----------------------------------------------------------

    //defer conn.Close()
}




    // Send data to the server
    // var i int = 0
    // for i < 3 {
    //      data := []byte("Hello, Server!")
    //      _, err = conn.Write(data)
    //      if err != nil {
    //         fmt.Println("Error:", err)
    //         return
    //      }
    //      time.Sleep(5 * time.Second)   
    //      i = i + 1 
    // }
 
