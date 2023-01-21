package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

const nodenum = 6

type send2node struct {
	Nodes  [nodenum]string
	Random string
}

func processconn(conn net.Conn) {
	defer conn.Close() // 关闭连接
	reader := bufio.NewReader(conn)
	fmt.Println(conn.RemoteAddr())
	var buf [12800]byte
	n, err := reader.Read(buf[:]) // 读取数据
	if err != nil {
		fmt.Println(err)
	}

	if err := os.WriteFile(conn.LocalAddr().String()+"nodes&RN.gob", buf[:n], 0666); err != nil {
		log.Fatal(err)
	}
	tempbyte, err := os.ReadFile(conn.LocalAddr().String() + "nodes&RN.gob")
	if err != nil {
		log.Fatal(err)
	}
	var temp send2node
	decoder := gob.NewDecoder(bytes.NewReader(tempbyte))
	decoder.Decode(&temp)

	fmt.Println("Received a new node's address ", temp)
	conn.Write([]byte("Got it")) // 发送数据
}

// TCP 客户端
func main() {
	addPtr := flag.String("addr", "127.0.0.1:7770", "Address")
	flag.Usage = func() {
		fmt.Println("Usage: [-addr string]")
		flag.PrintDefaults()
	}
	flag.Parse()
	//myaddress := "127.0.0.1:57267"
	myaddress := *addPtr
	addr, _ := net.ResolveTCPAddr("tcp", myaddress)
	hostaddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	conn, err := net.DialTCP("tcp", addr, hostaddr)
	if err != nil {
		fmt.Println("err : ", err)
		return
	}
	defer conn.Close()                                       // 关闭TCP连接
	_, err2 := conn.Write([]byte(conn.LocalAddr().String())) // 发送数据
	if err2 != nil {
		return
	}
	buf := [512]byte{}
	n, err := conn.Read(buf[:])

	if err != nil {
		fmt.Println("recv failed, err:", err)
		return
	}
	fmt.Println("have send local ip:", conn.LocalAddr())
	fmt.Println("trusted leader reply: ", string(buf[:n]))

	//receive nodes information that
	listen, err := net.Listen("tcp", conn.LocalAddr().String())
	if err != nil {
		fmt.Println("Listen() failed, err: ", err)
		return
	}
	for {
		conn2, err := listen.Accept() // 监听客户端的连接请求
		if err != nil {
			fmt.Println("Accept() failed, err: ", err)
		}
		go processconn(conn2) // 启动一个goroutine来处理客户端的连接请求
	}
}
