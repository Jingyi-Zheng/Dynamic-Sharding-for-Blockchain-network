package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
)

const nodenum = 6

type send2node struct {
	Nodes  [nodenum]string
	Random string
}

var mutex sync.Mutex
var nodes [nodenum]string
var now_nodesnum = -1

func get_Random() string {
	resp, err := http.Get("https://drand.cloudflare.com/8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce/public/latest")
	if err != nil {
		fmt.Println("error in get Drand")
		return "error in get Drand"
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	random := string(body)[31:95]
	return random
}

// TCP Server端测试
// 处理函数

func process(conn net.Conn) {
	if now_nodesnum == len(nodes)-1 {
		conn.Write([]byte("No space"))
		conn.Close()
		return
	}
	defer conn.Close() // 关闭连接
	reader := bufio.NewReader(conn)
	fmt.Println(conn.RemoteAddr())
	var buf [128]byte
	n, err := reader.Read(buf[:]) // 读取数据
	if err != nil {
		fmt.Println(err)
	}
	recvStr := string(buf[:n])
	fmt.Println("Received a new node's address ", recvStr)
	mutex.Lock()
	now_nodesnum += 1
	mutex.Unlock()
	nodes[now_nodesnum] = conn.RemoteAddr().String()
	conn.Write([]byte("Got it")) // 发送数据
}

func main() {
	hostaddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	listen, err := net.ListenTCP("tcp", hostaddr)
	if err != nil {
		fmt.Println("Listen() failed, err: ", err)
		return
	}

	for {
		if now_nodesnum == len(nodes)-1 {
			fmt.Println("Got all nodes' ip")
			fmt.Println(nodes)
			break
		}
		conn, err := listen.Accept() // 监听客户端的连接请求
		if err != nil {
			fmt.Println("Accept() failed, err: ", err)
			continue
		}
		go process(conn) // 启动一个goroutine来处理客户端的连接请求
		fmt.Println(now_nodesnum)
		fmt.Println(len(nodes))
	}

	sendMsg := send2node{nodes, get_Random()}
	fmt.Println(sendMsg)

	var TCPConn [len(nodes)]*net.TCPConn
	for i := 0; i < len(nodes); i++ {
		addr, err0 := net.ResolveTCPAddr("tcp", nodes[i])
		if err0 != nil {
			fmt.Println("err : ", err)
			return
		}
		conn2, err := net.DialTCP("tcp", nil, addr)
		TCPConn[i] = conn2
		if TCPConn[i] != nil {
			fmt.Println((TCPConn[i]))
		}
		if err != nil {
			fmt.Println("err : ", err)
			return
		}
	}
	sendByte := bytes.Buffer{}
	enc := gob.NewEncoder(&sendByte)
	err3 := enc.Encode(sendMsg)
	if err3 != nil {
		fmt.Println(err3)

	}
	fmt.Println(sendByte.Bytes())
	for i := 0; i < len(TCPConn); i++ {
		_, err2 := (TCPConn[i]).Write(sendByte.Bytes())
		if err2 != nil {
			return
		}
	}

}
