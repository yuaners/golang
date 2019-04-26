package netutil

import (
	"testing"
	"net"
	"fmt"
)

func TestUpdateConn(t *testing.T) {
	listener,err := net.Listen("tcp", ":9990")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("listen", listener.Addr().String())
	go func() {
		conn,err := listener.Accept()
		if err != nil {
			panic(err)
		}
		conn = UpdateConn(conn)
		go func(conn net.Conn) {
			defer conn.Close()
			var buf = make([]byte, 1024)
			for {
				n,err := conn.Read(buf)
				if err != nil {
					fmt.Println("read error:", err.Error())
					break
				}
				fmt.Println("read len:", n)
			}
		}(conn)
	}()

	conn,err := net.Dial("tcp", ":9990")
	if err != nil {
		panic(err)
	}
	fmt.Println("dial:", conn.LocalAddr().String())

	conn = UpdateConn(conn)
	defer conn.Close()

	for i := 0; i < 10; i++ {
		n,err := conn.Write([]byte(fmt.Sprintf("hello%03d", i)))
		if err != nil {
			panic(err)
		}
		fmt.Println(i, "send len:", n)
	}
}
