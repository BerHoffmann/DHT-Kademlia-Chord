package main

import (
    "fmt"
    "net"
    "time"
    "os"
    // "encoding/json"
    "strings"
    "strconv"
    "math"
    "math/rand"
)

var(
	hashSize = 6 // 2^6 nodes
    bucketSize = 2
    distanceToStore = 4
	newNode Node = Node{}
	myNode Node = Node{}
    distanceTable = [6][]Node{}
)


//o cliente, uma vez que entrou no CHORD, vai ter que fazer o mesmo papel que o server, ou seja, 
//ficar ouvindo na porta 8123, como em net.Listen("tcp", ":8123")

func sendString() {
    str := "conteudo da string"
    name := "teste.txt"
    hash := Hash([]byte(name), hashSize)
    fmt.Println("hash: ", hash)
    M := Message {
        RequestType : StoreData,
        Data : []byte(str),
        DataName : []byte(name),
        DataHash : hash,
    }
    jsonData := marshal(M)
    //send this message to this node
    conn, err := net.Dial("tcp", "182.19.0.3:8123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    _, err = conn.Write(jsonData)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
}

func search(name string) {
    hash := Hash([]byte(name), hashSize)
    fmt.Println("hash: ", hash)
    M := Message {
        RequestType : FindData,
        DataName : []byte(name),
        DataHash : hash,
        MarkedNodes : []Node{},
    }
    jsonData := marshal(M)
    //send this message to this node
    conn, err := net.Dial("tcp", "182.19.0.3:8123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    _, err = conn.Write(jsonData)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
}



func main() {

    listener, err := net.Listen("tcp", ":8123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    i, _ := strconv.Atoi(os.Args[2])
    time.Sleep(time.Duration(i) * time.Second)

    conn, err := net.Dial("tcp", "182.19.1."+ os.Args[1] + ":8123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    myNode.Addr = []byte(strings.Split(conn.LocalAddr().String(), ":")[0])
    myNode.HashPosition = Hash(myNode.Addr, hashSize)
    addInTable(myNode.HashPosition, myNode.Addr)
    
    M := Message {
        RequestType : AskForKNodes,
        Data : myNode.Addr,
        DataHash : myNode.HashPosition,
        RID : myNode.HashPosition,
    }

    jsonData := marshal(M)
	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
    
    HandleMessage(conn)

    //create random hash for each kbucket (from 2^0 to 2^6) and ask for nodes using this hash
    for i := 0; i < hashSize; i++ {
        //random hash is a random number between 2î and 2î+1
        randomHash := rand.Intn(int(math.Pow(2, float64(i+1)) - math.Pow(2, float64(i)))) + int(math.Pow(2, float64(i)))
        M = Message {
            RequestType : AskForKNodes,
            Data : myNode.Addr,
            DataHash : myNode.HashPosition,
            RID : randomHash,
        }
        ClosestNodes := FindClosestNodes(distanceTable, randomHash)
        //send this message to this nodes
        for j := 0; j < len(ClosestNodes); j++ {
            conn, err := net.Dial("tcp", string(ClosestNodes[j].Addr) + ":8123")
            if err != nil {
                fmt.Println("Error:", err)
                return
            }
            jsonData := marshal(M)
            _, err = conn.Write(jsonData)
            if err != nil {
                fmt.Println("Error:", err)
                return
            }
            HandleMessage(conn)
        }
    }

    // -------------------------------------------------------
    printDistanceTable()
    defer listener.Close()

	if myNode.HashPosition == 21{
        time.Sleep(time.Duration(10) * time.Second)
        sendString()
    }

    if myNode.HashPosition == 19{
        time.Sleep(time.Duration(20) * time.Second)
        search("teste.txt")
    }

    fmt.Println("Server is listening on port 8123")

 

    for {
        // Accept incoming connections
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error:", err)
            continue
        }

		go HandleMessage(conn)
	}
    
    // -----------------------------------------------------------
}



 
