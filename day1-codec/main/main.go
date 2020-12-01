package main

import (
	"encoding/json"
	"fmt"
	"geerpc"
	"geerpc/codec"
	"log"
	"net"
	"time"
)

/*
1、在 startServer 中使用了信道 addr，确保服务端端口监听成功，客户端再发起请求。
2、客户端首先发送 Option 进行协议交换，接下来发送消息头 h := &codec.Header{}，和消息体 geerpc req ${h.Seq}。
3、最后解析服务端的响应 reply，并打印出来。
*/

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}

	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

func main() {
	// 输出到标准输出
	log.SetFlags(log.LstdFlags)

	addr := make(chan string)

	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)

	defer func() {
		_ = conn.Close()
	}()

	time.Sleep(time.Second)
	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)

	cc := codec.NewGobCodec(conn)

	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           0,
		}
		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		_ = cc.ReadHeader(h)

		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
