package main

import (
	"io"
	"log"
	"net"
	"os"
)

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

func main() {
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// 读取并打印服务端的响应
	go mustCopy(os.Stdout, conn)
	// main goroutine从标准输入流中读取内容并将其发送给服务器
	mustCopy(conn, os.Stdin)
}
