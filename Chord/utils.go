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
	"sync"
	"strconv"
	"os"
	//"time"
	"bufio"
)

type Node struct {
	Addr []byte
	HashPosition int
}

type Data struct {
	FileName []string

}

//file sttruct containing data and dataname. hash by dataname

//essa struct vai ter que ter no package client tambem!
type Request string

type Message struct {
	RequestType Request
	Data []byte //ip or string
	DataName []byte
	DataHash int
	
	PrecessorAddr []byte
	PrecessorHash int
	SucessorAddr []byte
	SucessorHash int
}

type NodeConfig struct {
	mu sync.Mutex
	Precessor Node
	Sucessor Node
	MyNode Node
	DataList map[string]string 
}

var HashSize = 6.0

const(
	JoinRing Request = "request"
	SendMsg Request = "send"
	StoreMsg Request = "store"
	SearchMsg Request = "search"
	ResponseSearchMsg Request = "responseSearchMsg"
	ResultSearchMsg Request = "resultSearchMsg"
	SetNodes Request = "SetNodes"
	SetSuc Request = "setSuc"
	SetPrec Request = "setPrec"
	Print Request = "print"
)

func Hash(ip []byte, m float64) int {
	value := 0
	for _, v := range ip {
		value = value + int(v)*int(v)
	}

	rest := value % int(math.Pow(2, m))
	return rest
}

func (node * NodeConfig) RingIsUnitary() bool {
	return (node.Precessor.HashPosition == -1 && node.Precessor.HashPosition == -1)
}

func (node *NodeConfig) HandleMessage(conn net.Conn) {
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
	case JoinRing:
		node.HandleRequest(conn, M)
	case SetNodes:
		node.HandleSetNode(conn, M)
	case SetPrec:
		node.HandleSetPrec(conn, M)
	case SendMsg:
		node.HandleSendMsg(conn, M)
	case StoreMsg:
		node.HandleStoreMsg(conn, M)
	case SearchMsg:
		node.HandleSearchMsg(conn, M)
	case ResponseSearchMsg:
		node.HandleResponseSearchMsg(conn, M)
	case ResultSearchMsg:
		node.HandleResultSearchMsg(conn, M)
	case Print:
		node.PrintRing(conn, M)
	default:
		fmt.Println("tipo de request nao foi codado ainda!")
	}
}

func (node * NodeConfig) PrintContent() {
	fmt.Printf("My hash %d Data: %v\n", node.MyNode.HashPosition, node.DataList)
}

// func SendResponseSearchMsg(addrToSend, M Message) {

// }

func (node *NodeConfig) HandleSearchMsg(conn net.Conn, M Message) {
	conn.Close()
	if node.IsDataToMySucessor(M.DataHash) {
		// aqui, podemos apenas mudar o requesttype e entao usar pass along!
		// ficara entao menos codigo.
		//SendResponseSearchMsg()
		M.RequestType = ResponseSearchMsg
		PassAlongMsg(string(node.Sucessor.Addr), M)
	} else {
		PassAlongMsg(string(node.Sucessor.Addr), M)
	}
}

func (node * NodeConfig) HandleResponseSearchMsg(conn net.Conn, M Message) {
	conn.Close()
	conn, err := net.Dial("tcp", string(M.Data) + ":8123")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	value, exists := node.DataList[string(M.DataName)]

	if exists {
		M.RequestType = ResultSearchMsg
		M.Data = []byte( value )
		jsonData := marshal(M)

		_, err = conn.Write(jsonData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	} else {
		if string(M.Data) == string(node.MyNode.Addr) {
			fmt.Println("Search had no results")
			return 
		}
		M.RequestType = ResponseSearchMsg
		PassAlongMsg(string(node.Sucessor.Addr), M)
	}

}

func (node * NodeConfig) HandleResultSearchMsg(conn net.Conn, M Message) {
	conn.Close()
	// recebe msg e salva o arquivo em sua estrutura de dados dedicada!
	if node.DataList == nil {
		node.DataList = make(map[string]string)
	}
	
	node.DataList[string(M.DataName)] = string(M.Data)
	fmt.Println("Result of my search request(filename file):", string(M.DataName),
	node.DataList[string(M.DataName)])
}


func (node *NodeConfig) HandleStoreMsg(conn net.Conn, M Message) {
	conn.Close()
	if node.DataList == nil {
		node.DataList = make(map[string]string)
	}

	// eventualmente usar lock aqui!
	node.DataList[string(M.DataName)] = string(M.Data)
	node.PrintContent()
}

func (node *NodeConfig) HandleSendMsg(conn net.Conn, M Message)  {
	conn.Close()
	if node.IsDataToMySucessor(M.DataHash) {
		SendStoreMsg(string(node.Sucessor.Addr), M)
	} else {
		// repassar mesma mensagem ao sucessor
		PassAlongMsg(string(node.Sucessor.Addr), M)
	}
}

//melhor uma funcao encaminhar msg

func PassAlongMsg(addrToSend string, M Message) {
	conn, err := net.Dial("tcp", addrToSend + ":8123")
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }
	defer conn.Close()

	//M.RequestType = StoreMsg
	jsonData := marshal(M)

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }
}

func SendStoreMsg(addrToSend string, M Message) {
	conn, err := net.Dial("tcp", addrToSend + ":8123")
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }
	// nao precisa, pois o handlemsg->handlestore msg vai fechar conexao
	//defer conn.Close()

	M.RequestType = StoreMsg
	jsonData := marshal(M)

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }
}

func (node *NodeConfig) PrintConfig() {
	fmt.Println("My hash ",node.MyNode.HashPosition, "My prec and suc after update: ", string(node.Precessor.Addr),
	node.Precessor.HashPosition, string(node.Sucessor.Addr), node.Sucessor.HashPosition)
}


func (node *NodeConfig) HandleSetNode(conn net.Conn, M Message) {
	conn.Close()
	node.Precessor.Addr = M.PrecessorAddr
	node.Precessor.HashPosition = M.PrecessorHash
	node.Sucessor.Addr = M.SucessorAddr
	node.Sucessor.HashPosition = M.SucessorHash
	node.PrintConfig()
	
	// falar pro meu sucessor atualizar o endereço do seu antecessor
	M = Message{
		RequestType :   SetPrec,
		PrecessorAddr : node.MyNode.Addr,
		PrecessorHash : node.MyNode.HashPosition,
	}

	conn, err := net.Dial("tcp", string(node.Sucessor.Addr) + ":8123")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//defer conn.Close()

	jsonData := marshal(M)
	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func (node *NodeConfig) HandleRequest(conn net.Conn, M Message) {

	if node.RingIsUnitary() {
		node.Precessor.Addr = []byte(strings.Split(conn.RemoteAddr().String(), ":")[0])
		node.Precessor.HashPosition = Hash(node.Precessor.Addr, HashSize)
		node.Sucessor.Addr = []byte(strings.Split(conn.RemoteAddr().String(), ":")[0])
		node.Sucessor.HashPosition = Hash(node.Precessor.Addr, HashSize)
		node.MyNode.Addr = []byte(strings.Split(conn.LocalAddr().String(), ":")[0])
		node.MyNode.HashPosition = Hash(node.MyNode.Addr, HashSize)
		conn.Close()
		SendSetNodes(string(M.Data), node.MyNode.Addr, node.MyNode.Addr,
		node.MyNode.HashPosition, node.MyNode.HashPosition)
	} else {
		conn.Close()
		// verifico se é meu vizinho
		if node.IsMySuccessor(M.DataHash) {
			// mando pro nó a posição dele e atualizo o meu sucessor pra ele
			SendSetNodes(string(M.Data), node.Sucessor.Addr, node.MyNode.Addr, node.Sucessor.HashPosition, node.MyNode.HashPosition)
			node.Sucessor.Addr = M.Data
			node.Sucessor.HashPosition = M.DataHash
			node.PrintConfig()
		} else {
			// mando a mensagem pro meu sucessor
			PassAlongMsg(string(node.Sucessor.Addr), M)
			// conn, err := net.Dial("tcp", string(node.Sucessor.Addr) + ":8123")
			// if err != nil {
			// 	fmt.Println("Error:", err)
			// 	return
	 		// }
			// defer conn.Close()

			// jsonData := marshal(M)
			// _, err = conn.Write(jsonData)
			// if err != nil {
			// 	fmt.Println("Error:", err)
			// 	return
			//}
		}
	}

	// fmt.Println("precessor and sucessor", string(node.Precessor.Addr),
	// node.Precessor.HashPosition,
	// string(node.Sucessor.Addr), node.Sucessor.HashPosition)
	//time.Sleep(2 * time.Second)

}

func (node *NodeConfig) HandleSetPrec(conn net.Conn, M Message) {
	conn.Close()
	node.Precessor.Addr = M.PrecessorAddr
	node.Precessor.HashPosition = M.PrecessorHash
	node.PrintConfig()
}

func SendSetNodes(addrToSend string, sucAddr, precAddr []byte, sucHash, precHash int) {

	M := Message{
		RequestType :   SetNodes,
		PrecessorAddr : precAddr,
		SucessorAddr :  sucAddr,
		PrecessorHash : precHash,
		SucessorHash :  sucHash,
	}

	firstPart := addrToSend//strings.Split(addrToSend, ":")[0]
	conn, err := net.Dial("tcp", firstPart + ":8123")
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }
	//defer conn.Close()
	//fmt.Println("oi", M.RequestType, string(M.PrecessorAddr), M.PrecessorHash)
	
	jsonData := marshal(M)

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }

}

func (node *NodeConfig) IsMySuccessor(hash int) bool {

	if node.MyNode.HashPosition < hash {
		if node.Sucessor.HashPosition > node.MyNode.HashPosition && node.Sucessor.HashPosition > hash {
			return true
		}
		return false
	}

	if node.MyNode.HashPosition > hash {
		if node.Sucessor.HashPosition < node.MyNode.HashPosition && node.Sucessor.HashPosition > hash {
			return true
		}
		return false
	}

	return true
}

func (node *NodeConfig) IsDataToMySucessor(hash int) bool {
	if node.MyNode.HashPosition < hash {
		if node.Sucessor.HashPosition > node.MyNode.HashPosition && node.Sucessor.HashPosition >= hash {
			return true
		} else if node.Sucessor.HashPosition < node.MyNode.HashPosition {
			return true
		} else {
		return false
		}
	}

	if node.MyNode.HashPosition > hash {
		if node.Sucessor.HashPosition < node.MyNode.HashPosition && node.Sucessor.HashPosition >= hash {
			return true
		}
		return false
	}

	return true
}

func marshal(in any) []byte {
    out, err := json.Marshal(in)
    if err != nil {
        log.Fatalf("Unable to marshal due to %s\n", err)
    }

    return out
}

func (node *NodeConfig) HandleUserInput() {
	//var cmd string
	var reader *bufio.Reader
	var control bool

	if len(os.Args) == 2 {

		filename := os.Args[1]
		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error opening file:", err)
			os.Exit(1)
		}	

		reader = bufio.NewReader(file)
		control = false
	} else {
		reader = bufio.NewReader(os.Stdin)
		control = true
	}
	for {
		cmd, err := reader.ReadString('\n')
    
    	if err != nil {
			if control {
        		fmt.Println("Error reading input:", err)
	        	return
			} else {
				reader = bufio.NewReader(os.Stdin)
				control = true
			}
    	}

		// cmd usage: cmd [args ...]
		switch {
		case strings.HasPrefix(cmd, "print"): //print
			node.PrintConfig()
			node.PrintContent()

		case strings.HasPrefix(cmd, "join"): //join
			str := strings.Split(cmd, " ")[1]
			str = strings.TrimRight(str, "\n")
			node.SendJoinRing(str)

		case strings.HasPrefix(cmd, "send"): //search file
			dataname := strings.Split(cmd, " ")[1]
			data := strings.Split(cmd, " ")[2]
			data = strings.TrimRight(data, "\n")
			node.SendMessage(dataname, data)

		case strings.HasPrefix(cmd, "search"):
			str := strings.Split(cmd, " ")[1]
			str = strings.TrimRight(str, "\n")
			fmt.Println("Search request for file", str)
			node.SearchMessage(str)

		case strings.HasPrefix(cmd, "sleep"):
			tim := strings.Split(cmd, " ")[1]
			tim = strings.TrimRight(tim, "\n")
			t, _ := strconv.Atoi(tim)
			//time.Sleep(time.Duration(i + 2) * time.Second)
			time.Sleep( time.Duration(t) * time.Second)

		case strings.HasPrefix(cmd, "ring"):
			M := Message{
				RequestType : Print,
				Data : node.MyNode.Addr,
			}
			conn, err := net.Dial("tcp", string(node.Sucessor.Addr)+":8123")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			node.PrintRing(conn, M)
		case len([]byte(cmd)) == 0:
		default:
			fmt.Println("\nTry again, sent", []byte(cmd))
		}


	}
}

func(node *NodeConfig) PrintRing(conn net.Conn, M Message) {
	conn.Close()
	if string(node.Sucessor.Addr) == string(M.Data) {
		fmt.Printf("-> (%d %s)", node.MyNode.HashPosition, string(node.MyNode.Addr))
		fmt.Printf("-> (%d %s) ->", node.Sucessor.HashPosition, string(node.Sucessor.Addr))
	} else {
		fmt.Printf("-> (%d %s)", node.MyNode.HashPosition, string(node.MyNode.Addr))
		PassAlongMsg(string(node.Sucessor.Addr), M)
	}
}

//go has a convention of using first uppercase if i want to export this
//variables outside this file. functions too!
//exception: in 


func (N *NodeConfig) SendJoinRing(addrToSend string) {
	conn, err := net.Dial("tcp", addrToSend + ":8123")
    N.MyNode.Addr = []byte(strings.Split(conn.LocalAddr().String(), ":")[0])
    N.MyNode.HashPosition = Hash(N.MyNode.Addr, HashSize)
    
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    M := Message {
        RequestType : JoinRing,
        Data : []byte(strings.Split(conn.LocalAddr().String(), ":")[0]),
        DataHash : Hash([]byte(strings.Split(conn.LocalAddr().String(), ":")[0]), 
        HashSize),
    }

    jsonData := marshal(M)

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func (node * NodeConfig) SendMessage(name, data string) {
    M := Message{
        RequestType: SendMsg,
        Data: []byte(data),
        DataName: []byte(name), 
        DataHash: Hash([]byte(name), HashSize),
    }

	//fmt.Println("quem vou conectar: ", string(node.Sucessor.Addr))
    conn, err := net.Dial("tcp", string(node.Sucessor.Addr) + ":8123")

    jsonData := marshal(M)
	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }
}

func (node * NodeConfig) SearchMessage(name string)  {

    M := Message{
        Data: node.MyNode.Addr,
        RequestType: SearchMsg,
        DataName: []byte(name), 
        DataHash: Hash([]byte(name), HashSize),
    }

    // i, _ := strconv.Atoi(os.Args[2])
    // time.Sleep(time.Duration(i + 2) * time.Second)
    conn, err := net.Dial("tcp", string(node.Sucessor.Addr) + ":8123")

    jsonData := marshal(M)
	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	 }

}



// func (n *NodeConfig) handleMessage(conn net.Conn) {

//     defer conn.Close()
//     var M Message
//     buf := make([]byte, 1024)   
//     for { 
//         time.Sleep(1 * time.Second)
//         _, err := conn.Read(buf)
//         if err != nil {
//             fmt.Println("oi")
//             fmt.Println("Error:", err)
//            return
//         }

//         err = json.Unmarshal(buf, &M)
//         if err != nil {
//             fmt.Println("oi2")
//             fmt.Println("Error:", err)
//            return
//         }

// 		n.Precessor.Addr = M.PrecessorAddr
// 		n.Precessor.HashPosition = M.PrecessorHash
// 		n.Sucessor.Addr = M.SucessorAddr
// 	    n.Sucessor.HashPosition = M.SucessorHash
//         n.MyNode.Addr = []byte(conn.LocalAddr().String())
// 		n.MyNode.HashPosition = Hash(n.MyNode.Addr, HashSize)

//         fmt.Println("precessor and sucessor", string(n.Precessor.Addr),
// 	    n.Precessor.HashPosition,
// 	    string(n.Sucessor.Addr), n.Sucessor.HashPosition)
//     }
//     time.Sleep(1 * time.Second)
// }