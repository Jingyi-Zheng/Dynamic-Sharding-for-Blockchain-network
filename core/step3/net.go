package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type nodelist_shard struct {
	Nodes              [Nodenum]string
	FirstShard         [Nodenum]int
	SecondShard        [Nodenum]int
	FirstShard_leader  [Nodenum]int
	SecondShard_leader [Nodenum]int
}
type nodes_inf struct {
	NodeAddress     string
	First_shard     int
	Second_shard    int
	Beleader_level1 int
	Beleader_level2 int
}
type ConnectionsQueue chan string

type NodeChannel chan *Node

type Node struct {
	*net.TCPConn
	lastSeentime int
}

type Nodes map[string]*Node

type Network struct {
	Nodes
	Nodes_level2              Nodes
	Nodes_level1              Nodes
	ConnectionsQueue_level1   ConnectionsQueue
	ConnectionsQueue_level2   ConnectionsQueue
	ConnectionCallback_level1 NodeChannel
	ConnectionCallback_level2 NodeChannel
	//节点字符串和tcp conn的map
	Address string
	//本地address
	BroadcastQueue_level1 chan Message
	BroadcastQueue_level2 chan Message
	//待发送队列
	IncomingMessages chan Message
	//收到消息队列
}

func Readnodesdb() nodelist_shard {
	tempbyte, err := os.ReadFile("../step2/nodes_shard.gob")
	if err != nil {
		log.Fatal(err)
	}
	var nodesdb nodelist_shard
	decoder := gob.NewDecoder(bytes.NewReader(tempbyte))
	decoder.Decode(&nodesdb)
	return nodesdb
}

func readNodesinfo(myaddress string) ([]nodes_inf, []nodes_inf) {
	var myshard1 int
	var myshard2 int
	nodesdb := Readnodesdb()
	var node_level1 []nodes_inf
	var node_level2 []nodes_inf
	for i := 0; i < Nodenum; i++ {
		if nodesdb.Nodes[i] == myaddress {
			myshard1 = nodesdb.FirstShard[i]
			myshard2 = nodesdb.SecondShard[i]
			break
		}
	}
	for j := 0; j < Nodenum; j++ {
		if nodesdb.FirstShard[j] == myshard1 {
			node_level1 = append(node_level1, nodes_inf{nodesdb.Nodes[j], nodesdb.FirstShard[j], nodesdb.SecondShard[j], nodesdb.FirstShard_leader[j], nodesdb.SecondShard_leader[j]})
		}
		if nodesdb.SecondShard[j] == myshard2 && nodesdb.FirstShard[j] != myshard1 {
			node_level2 = append(node_level2, nodes_inf{nodesdb.Nodes[j], nodesdb.FirstShard[j], nodesdb.SecondShard[j], nodesdb.FirstShard_leader[j], nodesdb.SecondShard_leader[j]})
		}
	}
	return node_level1, node_level2
}

func SetupNetwork(address string) *Network {
	fmt.Println("set up in:", address)
	n := new(Network)
	n.BroadcastQueue_level1, n.BroadcastQueue_level2, n.IncomingMessages = make(chan Message), make(chan Message), make(chan Message)
	n.ConnectionsQueue_level1, n.ConnectionsQueue_level2, n.ConnectionCallback_level1, n.ConnectionCallback_level2 = CreateConnectionsQueue(address)
	n.Nodes_level1 = Nodes{}
	n.Nodes_level2 = Nodes{}
	n.Nodes = Nodes{}
	n.Address = address
	return n
}

func CreateConnectionsQueue(myaddress string) (ConnectionsQueue, ConnectionsQueue, NodeChannel, NodeChannel) {
	in1 := make(ConnectionsQueue)
	in2 := make(ConnectionsQueue)
	out1 := make(NodeChannel)
	out2 := make(NodeChannel)

	go func() {
		for {
			address1 := <-in1

			if address1 != myaddress {

				go ConnectToNode(address1, 5*time.Second, true, out1)
			}
		}
	}()
	go func() {
		for {
			address2 := <-in2

			if address2 != myaddress {

				go ConnectToNode(address2, 5*time.Second, true, out2)
			}
		}
	}()

	return in1, in2, out1, out2
}
func ConnectToNode(dst string, timeout time.Duration, retry bool, cb NodeChannel) {
	addrDst, err := net.ResolveTCPAddr("tcp4", dst)
	networkError(err)
	var con *net.TCPConn = nil
loop:
	for {

		breakChannel := make(chan bool)
		go func() {
			con, err = net.DialTCP("tcp", nil, addrDst)
			if con != nil {
				fmt.Println("has connected to", con.RemoteAddr())
				cb <- &Node{con, int(time.Now().Unix())}
				breakChannel <- true
			}
		}()

		select {
		case <-Timeout(timeout):
			if !retry {
				break loop
			}
		case <-breakChannel:
			break loop
		}

	}
}

func (n Nodes) AddNode(node *Node, network Network) bool {

	key := node.TCPConn.RemoteAddr().String()

	if key != network.Address && n[key] == nil {

		n[key] = node

		go HandleNode(node, network)

		return true
	}
	return false
}

func (n *Network) Run() {

	fmt.Println("Listening in", n.Address)
	listencb := StartListening(n.Address, *n)

	for {
		select {
		case node := <-listencb:
			n.Nodes.AddNode(node, *n)
		case node := <-n.ConnectionCallback_level1:
			n.Nodes_level1.AddNode(node, *n)
		case node := <-n.ConnectionCallback_level2:
			n.Nodes_level2.AddNode(node, *n)
		case message1 := <-n.BroadcastQueue_level1:
			go n.BroadcastMessage_level1(message1)
		case message2 := <-n.BroadcastQueue_level2:
			go n.BroadcastMessage_level2(message2)
		}
	}
}

func (n *Network) BroadcastMessage_level1(message Message) {

	b, _ := message.Message2Bytes()

	for k, node := range n.Nodes_level1 {
		fmt.Println(node.LocalAddr(), "Broadcasting to", k)
		go func() {
			_, err := node.TCPConn.Write(b)
			if err != nil {
				fmt.Println("Error bcing to", node.TCPConn.RemoteAddr())
			}
		}()
	}
}

func (n *Network) BroadcastMessage_level2(message Message) {

	b, _ := message.Message2Bytes()

	for k, node := range n.Nodes_level2 {
		fmt.Println(node.LocalAddr(), "Broadcasting to", k)
		go func() {
			_, err := node.TCPConn.Write(b)
			if err != nil {
				fmt.Println("Error bcing to", node.TCPConn.RemoteAddr())
			}
		}()
	}
}

func StartListening(address string, network Network) NodeChannel {

	cb := make(NodeChannel)
	addr, err := net.ResolveTCPAddr("tcp4", address)
	networkError(err)

	listener, err := net.ListenTCP("tcp4", addr)
	networkError(err)

	go func(l *net.TCPListener) {
		for {
			connection, err := l.AcceptTCP()
			networkError(err)
			nodeacc := &Node{connection, int(time.Now().Unix())}
			cb <- nodeacc
		}
	}(listener)
	return cb
}
func HandleNode(node *Node, network Network) {

	for {
		var bs []byte = make([]byte, 1024*1000)
		n, err := node.TCPConn.Read(bs[0:])
		networkError(err)

		if err == io.EOF {
			fmt.Println("EOF")
			//remove the error
			node.TCPConn.Close()
			break
		}

		m := new(Message)
		err = m.Bytes2Message(bs[0:n])
		fmt.Println("receive from", node.RemoteAddr())
		if err != nil {
			fmt.Println(err)
			continue
		}

		m.Reply = make(chan Message)

		go func(cb chan Message) {
			for {
				m, ok := <-cb

				if !ok {
					close(cb)
					break
				}

				b, _ := m.Message2Bytes()
				l := len(b)

				i := 0
				for i < l {

					a, _ := node.TCPConn.Write(b[i:])
					i += a
				}
			}

		}(m.Reply)

		network.IncomingMessages <- *m
	}
}

func networkError(err error) {

	if err != nil && err != io.EOF {

		log.Println("Blockchain network: ", err)
	}
}

func (n *Network) connecttonode(node_info1 []nodes_inf, node_info2 []nodes_inf) {
	for i := 0; i < len(node_info1); i++ {
		n.ConnectionsQueue_level1 <- node_info1[i].NodeAddress
	}
	for j := 0; j < len(node_info2); j++ {
		n.ConnectionsQueue_level2 <- node_info2[j].NodeAddress
	}
}
func main() {

	addPtr := flag.String("addr", "127.0.0.1:7770", "Address")
	flag.Usage = func() {
		fmt.Println("Usage: [-addr string]")
		flag.PrintDefaults()
	}
	flag.Parse()

	myaddress := *addPtr

	mynetwork := SetupNetwork(myaddress)
	go func() {
		mynetwork.Run()
	}()
	time.Sleep(30 * time.Second)
	node_info1, node_info2 := readNodesinfo("127.0.0.1:7770")
	mynetwork.connecttonode(node_info1, node_info2)
	time.Sleep(15 * time.Second)
	mynetwork.BroadcastQueue_level1 <- *NewMessage(byte(1))
	mynetwork.BroadcastQueue_level2 <- *NewMessage(byte(1))
	for {
		select {
		case temp := <-mynetwork.IncomingMessages:
			fmt.Println(temp)

		}
	}
}
