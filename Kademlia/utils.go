package main

import(
	"math"
	"fmt"
	"encoding/json"
	"net"
	"time"
	"log"
	"bytes"
	"strings"
)

type Node struct {
	Addr []byte
	HashPosition int
	DataStorage [][]byte
}

type Data struct {
	FileName []string

}

//essa struct vai ter que ter no package client tambem!
type Request string

type Message struct {
	RequestType Request
	Data []byte //ip or string
	DataName []byte
	DataHash int
	MarkedNodes []Node
	ClosestNodes []Node
	RID int
}

const(
	SendMsg Request = "send"
	SearchMsg Request = "search"

	AskForKNodes Request = "askForKNodes"
	SendKNodes Request = "sendKNodes"
	Ping Request = "ping"
	PingResponse Request = "pingResponse"
	StoreData Request = "storeData"
	FindData Request = "findData"
	FindKNodes Request = "findKNodes"
	ResponseFindKNodes Request = "responseFindKNodes"
	YouStoreData Request = "youStoreData"
)

func addInTable(newNodeHash int, newNodeAddr []byte) {
	distance := Distance(newNodeHash, myNode.HashPosition)

	//  if dist is 0, kbucket_index is 0

	if distance == 0 {
		if len(distanceTable[0]) == 0 {
			distanceTable[0] = append(distanceTable[0], Node{newNodeAddr, newNodeHash, [][]byte{}})
		}
		return
	}
	
	kbucket_index := int(math.Log2(float64(distance)))
	//fmt.Println("Distance: ", distance, " kbucket_index: ", kbucket_index)

	newNode := Node{newNodeAddr, newNodeHash, [][]byte{}}
	if len(distanceTable[kbucket_index]) < bucketSize {
		// if hash not already in table
		for i := 0; i < len(distanceTable[kbucket_index]); i++ {
			if distanceTable[kbucket_index][i].HashPosition == newNodeHash{
				return
			}
		}
		distanceTable[kbucket_index] = append(distanceTable[kbucket_index], newNode)
	}
}

func Distance(hash1, hash2 int) int {
	return hash1 ^ hash2
}

func distanceTableUnitary() bool {
	return (len(distanceTable[0]) == 0)
}

// function that receives a distanceTable and a Hash and returt the k nodes closest to that hash
func FindClosestNodes(distanceTable [6][]Node, hash int) []Node {
	var closestNodes []Node
	var flag = false
	for i := 0; i < len(distanceTable); i++ {
		for j := 0; j < len(distanceTable[i]); j++ {
			flag = false
			if i == 0 && j == 0{
				continue
			}
			if (len(closestNodes) < bucketSize && distanceTable[i][j].HashPosition != hash) {
				closestNodes = append(closestNodes, distanceTable[i][j])
			} else {
				for k := 0; k < len(closestNodes); k++ {
					if Distance(hash, distanceTable[i][j].HashPosition) < Distance(hash, closestNodes[k].HashPosition) && distanceTable[i][j].HashPosition != hash {
						closestNodes[k] = distanceTable[i][j]
						flag = true
					}
					if flag == true {
						break
					}

				}
			}

		}
	}
	return closestNodes
}


func Hash(ip []byte, m int) int {
	
	value := 0
	for _, v := range ip {
		value = value + int(v)*int(v)
	}

	rest := value % int(math.Pow(2, float64(m)))
	return rest
}


func HandleMessage(conn net.Conn) {
	
	var M Message
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error:", err)
	   return
	}

	cleanbuf := bytes.Trim(buf, "\x00")
	err = json.Unmarshal(cleanbuf, &M)
	if err != nil {
		fmt.Println("Unmarshal error:", err)
		return
	}

	switch M.RequestType {
	case AskForKNodes:
		HandleRequest(conn, M)
	case SendKNodes:
		HandleSendKNodes(conn, M)
	case Ping:
		HandlePing(conn, M)
	case PingResponse:
		HandlePingResponse(conn, M)
	case StoreData:
		HandleStoreData(conn, M)
	case FindData:
		HandleFindData(conn, M)
	case FindKNodes:
		HandleFindKtNodesInTable(conn, M)
	case ResponseFindKNodes:
		HandleResponseFindKtNodesInTable(conn, M)
	case YouStoreData:
		HandleYouStoreData(conn, M)
	default:
		fmt.Println("tipo de request nao foi codado ainda!")
	}
}

func HandleRequest(conn net.Conn, M Message) {
	if distanceTableUnitary(){
		myNode.Addr = []byte(strings.Split(conn.LocalAddr().String(), ":")[0])
    	myNode.HashPosition = Hash(myNode.Addr, hashSize)
    	addInTable(myNode.HashPosition, myNode.Addr)
	}

	addInTable(M.DataHash, M.Data)

	ClosestNodes := FindClosestNodes(distanceTable, M.RID)

	M = Message{
		RequestType :  SendKNodes,
		Data : myNode.Addr,
		DataHash : myNode.HashPosition,
		ClosestNodes : ClosestNodes,
	}

	jsonData := marshal(M)

	_, err := conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	time.Sleep(3 * time.Second)
}


func printDistanceTable() {
	fmt.Println("Distance Table:")
	for i := 0; i < len(distanceTable); i++ {
		fmt.Println("Bucket ", i, ":")
		for j := 0; j < len(distanceTable[i]); j++ {
			fmt.Println("Node ", j, ": ", string(distanceTable[i][j].Addr), " ", distanceTable[i][j].HashPosition)
		}
	}
}

func HandleSendKNodes(conn net.Conn, M Message) {

	addInTable(M.DataHash, M.Data)		
	// print data hash and data
	//printDistanceTable()

	Ping_M := Message{
		RequestType : Ping,
		Data : myNode.Addr,
		DataHash : myNode.HashPosition,
	}

	conn.Close()

	for i := 0; i < len(M.ClosestNodes); i++ {
		conn, err := net.Dial("tcp", string(M.ClosestNodes[i].Addr) + ":8123")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		jsonData := marshal(Ping_M)
		_, err = conn.Write(jsonData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		go HandleMessage(conn)
	}
}

func HandlePing(conn net.Conn, M Message) {
	// fmt.Println("\nHandle Ping:")

	addInTable(M.DataHash, M.Data)
	// printDistanceTable()
	// send ping response with data as my IP
	M = Message{
		RequestType : PingResponse,
		Data : myNode.Addr,
		DataHash : myNode.HashPosition,
	}
	
	jsonData := marshal(M)
	_, err := conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//printDistanceTable()
}

func HandlePingResponse(conn net.Conn, M Message) {
	// fmt.Println("\nHandle Ping Response:")
	addInTable(M.DataHash, M.Data)
	// printDistanceTable()
	conn.Close()
	
}


func GetKMinorsFromArrays(arr1, arr2 []Node, hash, length int) []Node {
	// make temp vector with all elements of arr1 and arr2
	// if arr is empty
	if len(arr1) == 0 {
		return arr2
	}
	if len(arr2) == 0 {
		return arr1
	}

	var temp []Node
	for i := 0; i < len(arr1); i++ {
		temp = append(temp, arr1[i])
	}
	for i := 0; i < len(arr2); i++ {
		temp = append(temp, arr2[i])
	}

	var kMinors []Node
	var Min = 1000000
	var Min_index, Min2_index int
	for i := 0; i < len(temp); i++ {
		if Distance(hash, temp[i].HashPosition) < Min {
			Min = Distance(hash, temp[i].HashPosition)
			Min_index = i
		}
	}
	kMinors = append(kMinors,  temp[Min_index])
	var Min2 = 1000000
	for i := 0; i < len(temp); i++ {
		if Distance(hash, temp[i].HashPosition) < Min2 && Distance(hash, temp[i].HashPosition) > Min {
			Min2 = Distance(hash, temp[i].HashPosition)
			Min2_index = i
		}
	}
	kMinors = append(kMinors, temp[Min2_index])
	return kMinors
}

func CompareArrays(arr1, arr2 []Node) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i := 0; i < len(arr1); i++ {
		if arr1[i].HashPosition != arr2[i].HashPosition {
			return false
		}
	}
	return true
}

// find the k closest nodes in all hash table recursevely
func HandleFindKtNodesInTable(conn net.Conn, M Message) {
	myClosestNodes := FindClosestNodes(distanceTable, M.RID)
	allClosestNodes := GetKMinorsFromArrays(myClosestNodes, M.ClosestNodes, M.RID, len(myClosestNodes))

	if CompareArrays(allClosestNodes, M.ClosestNodes) {
		Response_M := Message{
			RequestType : ResponseFindKNodes,
			Data : M.Data,
			DataHash : M.DataHash,
			DataName : M.DataName,
			ClosestNodes : allClosestNodes,
		}
		jsonData := marshal(Response_M)
		_, err := conn.Write(jsonData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	} else {
		// send message to myClosestNodes
		M = Message{
			RequestType : FindKNodes,
			Data : M.Data,
			DataHash : M.DataHash,
			DataName : M.DataName,
			ClosestNodes : allClosestNodes,
			RID : M.RID,
		}

		// create array of node_connn of len(myClosestNodes)
		node_conn := make([]net.Conn, len(myClosestNodes))
		var err error
		for i := 0; i < len(myClosestNodes); i++ {
			node_conn[i], err = net.Dial("tcp", string(myClosestNodes[i].Addr) + ":8123")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			jsonData := marshal(M)
			_, err = node_conn[i].Write(jsonData)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
		// arr is a slice of slice of nodes
		arr := make([][]Node, len(myClosestNodes))
		for i := 0; i < len(myClosestNodes); i++ {
			arr[i] = HandleRespondeFromRecursion(node_conn[i])
		}
		response_closestNodes := GetKMinorsFromArrays(arr[0], arr[1], M.RID, len(myClosestNodes))

		Response_M := Message{
			RequestType : ResponseFindKNodes,
			Data : M.Data,
			DataHash : M.DataHash,
			DataName : M.DataName,
			ClosestNodes : response_closestNodes,
		}
		jsonData := marshal(Response_M)
		_, err = conn.Write(jsonData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
}

func HandleRespondeFromRecursion(conn net.Conn) []Node{

	var M Message
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error:", err)
	   return nil
	}

	cleanbuf := bytes.Trim(buf, "\x00")
	err = json.Unmarshal(cleanbuf, &M)
	if err != nil {
		fmt.Println("Unmarshal error:", err)
		return nil
	}

	if M.RequestType == ResponseFindKNodes {
		return M.ClosestNodes
	}
	return nil
}

func HandleResponseFindKtNodesInTable(conn net.Conn, M Message){
	nodes := M.ClosestNodes
	// for each of those closest nodes, send a message for them to store the data
	for i := 0; i < len(nodes); i++ {
		conn, err := net.Dial("tcp", string(nodes[i].Addr) + ":8123")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		M = Message{
			RequestType : YouStoreData,
			Data : M.Data,
			DataHash : M.DataHash,
			DataName : M.DataName,
			ClosestNodes : M.ClosestNodes,
		}
		jsonData := marshal(M)
		_, err = conn.Write(jsonData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
}

func HandleYouStoreData(conn net.Conn, M Message) {
	// check if my node has the data
	// for i := 0; i < len(myNode.DataStorage); i++ {
	// 	if bytes.Equal(myNode.DataStorage[i], M.DataName) {
	// 		// print data
	// 		fmt.Println("Data: ", string(myNode.DataStorage[i]))
	// 		return
	// 	}
	// }
	// if not, store it
	myNode.DataStorage = append(myNode.DataStorage, M.DataName)
	fmt.Println("Saving Data: ", string(myNode.DataStorage[0]))

}

func marshal(in any) []byte {
    out, err := json.Marshal(in)
    if err != nil {
        log.Fatalf("Unable to marshal due to %s\n", err)
    }

    return out
}

func HandleStoreData(conn net.Conn, M Message) {
	conn.Close()
	//check if my node has the data

	// distance := Distance(M.DataHash, myNode.HashPosition)
	ClosestNodes := FindClosestNodes(distanceTable, M.DataHash)

	
	conn, err := net.Dial("tcp", string(ClosestNodes[0].Addr) + ":8123")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	M = Message{
		RequestType : FindKNodes,
		Data : M.Data,
		DataName : M.DataName,
		DataHash : M.DataHash,
		ClosestNodes : ClosestNodes,
		RID : M.DataHash,
	}

	jsonData := marshal(M)
	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	HandleMessage(conn)
}

func HandleFindData(conn net.Conn, M Message) {

	//if my node is in the closest nodes, close the connection
	for i := 0; i < len(M.MarkedNodes); i++ {
		//compare hash
		if M.MarkedNodes[i].HashPosition == myNode.HashPosition {
			conn.Close()
			return
		}
	}

	fmt.Println("Finding Data...")
	defer conn.Close()
	//check if my node has the data
	for i := 0; i < len(myNode.DataStorage); i++ {
		if M.DataHash == Hash(myNode.DataStorage[i], hashSize) {
			// print data
			fmt.Println("Found Data: ", string(myNode.DataStorage[i]))
			return
		}
	}
	// distance := Distance(M.DataHash, myNode.HashPosition)
	MarkedNodes := FindClosestNodes(distanceTable, M.DataHash)
	//calculate distance to each of the closest nodes
	for i := 0; i < len(MarkedNodes); i++ {
		conn, err := net.Dial("tcp", string(MarkedNodes[i].Addr) + ":8123")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		M.MarkedNodes = append(M.MarkedNodes, myNode)
		jsonData := marshal(M)
		_, err = conn.Write(jsonData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
}