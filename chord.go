package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"unicode"
)

type NodeIdentifier struct {
	Id      string
	Address string
}

type ChordNode struct {
	nId         NodeIdentifier
	FingerTable []string
	Successors  []NodeIdentifier
	Predecessor NodeIdentifier

	Bucket map[string]string

	n_succesors int
	ts          int
	tff         int
	tcp         int
	mutex       sync.Mutex
}

func main() {

	ip := ""
	port := -1
	chord_ip := ""
	chord_port := -1
	ts := 1
	tff := 1
	tcp := 1
	n_successors := 1
	id := ""

	for i := 1; i < len(os.Args); i += 2 {
		fmt.Println("arg: ", os.Args[i], ", parameter: ", os.Args[i+1])
		var err error = nil
		switch os.Args[i] {
		case "-a":
			ip = os.Args[i+1]
		case "-p":
			port, err = strconv.Atoi(os.Args[i+1])
			if port < 0 || port > 65535 {
				fmt.Println("invalid port argument: ", port)
				return
			}
		case "--ja":
			chord_ip = os.Args[i+1]
		case "--jp":
			chord_port, err = strconv.Atoi(os.Args[i+1])
			if chord_port < 0 || chord_port > 65535 {
				fmt.Println("invalid chord_port argument: ", chord_port)
				return
			}
		case "--ts":
			ts, err = strconv.Atoi(os.Args[i+1])
			if ts < 1 || ts > 60000 {
				fmt.Println("invalid ts argument: ", ts)
				return
			}
		case "--tff":
			tff, err = strconv.Atoi(os.Args[i+1])
			if tff < 1 || tff > 60000 {
				fmt.Println("invalid tff argument: ", tff)
				return
			}
		case "--tcp":
			tcp, err = strconv.Atoi(os.Args[i+1])
			if tcp < 1 || tcp > 60000 {
				fmt.Println("invalid tcp argument: ", tcp)
				return
			}
		case "-r":
			n_successors, err = strconv.Atoi(os.Args[i+1])
			if n_successors < 1 || n_successors > 32 {
				fmt.Println("invalid r argument: ", n_successors)
				return
			}
		case "-i":
			id = os.Args[i+1]
			notValid := false
			for _, l := range id {
				if !unicode.IsLetter(l) || !unicode.IsDigit(l) {
					notValid = true
				}
			}
			if notValid || len(id) != 40 {
				fmt.Println("id should be 40 characters of [0-9a-fA-F]: ", id)
				return
			}
		}

		if err != nil {
			fmt.Println("exception when formating argument: ", os.Args[i], " value is:", os.Args[i+1])
			return

		}

	}

	if ip == "" || port == -1 {
		fmt.Println("-a and -p must be specified: ", ip, ":", port)
		return
	}

	if (chord_port != -1 && chord_ip == "") || (chord_port == -1 && chord_ip != "") {
		fmt.Println("if either jp or ja is specified the other must be specified: ", chord_ip, ":", chord_port)
	}

	port_string := strconv.Itoa(port)

	node := CreateChord(ip, port_string, n_successors, ts, tff, tcp)

	node.server()

	call("127.0.0.1:8080", "ChordNode.PrintState", &Empty{confirm: true}, &Empty{confirm: true})
}

func CreateChord(ip string, port string, n_successors int, ts int, tff int, tcp int) *ChordNode {

	address := ip + ":" + port
	// id := hashString(ip + ":" + port)
	id := ip + ":" + port
	node := ChordNode{ts: ts, tff: tff, tcp: tcp, n_succesors: n_successors, nId: NodeIdentifier{Id: id, Address: address}}

	node.FingerTable = []string{}
	node.Bucket = map[string]string{}
	node.Successors = make([]NodeIdentifier, n_successors)
	for i, _ := range node.Successors {
		node.Successors[i] = node.nId
	}
	node.Predecessor = node.nId

	return &node
}

func (node *ChordNode) server() error {
	rpc.Register(node)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", node.nId.Address)
	if e != nil {
		fmt.Println("unable to start node rpc server: ", e)
		return e
	}

	go http.Serve(l, nil)

	return nil
}

func call(address string, rpcname string, args interface{}, reply interface{}) bool {

	c, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		fmt.Println("dialing error:", err)
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println("calling error: ", err)
	return false
}

// func (n *ChordNode) JoinChord(Address string) {
//	for {

//		req := SuccFind{Id: n.nId.Id}
//		rep := Bingo{}
//		call(Address, &req, &rep)
//		n.Successor[0] = rep.SuccId
//		if rep.Identified {
//			break
//		}
//
//	}

//}

func (n *ChordNode) Put(args *Put, reply *Put_reply) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	fmt.Println("put: ", args.Id, " ", args.Value)

	n.Bucket[args.Id] = args.Value // security issue?
	reply.Confirm = true
	return nil
}

func (n *ChordNode) Get(args *Get, reply *Get_reply) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	fmt.Println("get: ", args.Id)
	reply.Confirm = false
	val, ok := n.Bucket[args.Id]
	if !ok {
		reply.Confirm = true
		reply.Content = val
	}

	return nil // security issue?
}

func (n *ChordNode) Delete(args *Delete, reply *Delete_reply) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	fmt.Println("delete: ", args.Id)
	reply.Confirm = false
	_, ok := n.Bucket[args.Id]
	if !ok {
		delete(n.Bucket, args.Id)
		reply.Confirm = true
	}

	return nil
}

// func Find_Successor(id) {
// 	fmt.Println("Find_Successor: ", id)

// }

// func (n *ChordNode) closest_preceding_node(id) {
// 	fmt.Println("closest_preceding_node: ", id)

// }

// func Lookup() {

// }

// func StoreFile(n *ChordNode, fileName string) {

// }

func (n *ChordNode) PrintState(args *Empty, reply *Empty) {
	fmt.Println("Node:")
	fmt.Println("Address: ", n.nId.Address, " Id: ", n.nId.Id)
	fmt.Println("Predecessor: ", n.Predecessor)
	fmt.Println("Fingertable")
	for i, finger := range n.FingerTable {
		fmt.Println("	", i, ": ", finger)
	}
	fmt.Println("Successors: ")
	for i, sucessor := range n.Successors {
		fmt.Println("	", i, ": ", sucessor)
	}
	fmt.Println("Bucket: ")
	for key, value := range n.Bucket {
		fmt.Println("	", key, ": ", value)
	}
	fmt.Println("n_succesors: ", n.n_succesors)
	fmt.Println("ts: ", n.ts)
	fmt.Println("tff: ", n.tff)
	fmt.Println("tcp: ", n.tcp)
}

// func hashString(elt string) *big.Int {
// 	hasher := crypto.SHA1.New()
// 	hasher.Write([]byte(elt))
// 	return new(big.Int).SetBytes(hasher.Sum(nil))
// }
