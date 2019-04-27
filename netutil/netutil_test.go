package netutil

import (
	"testing"
	"net"
	"fmt"
	"time"
	"sync"
)

func TestUpdateConn(t *testing.T) {
	listener,err := Listen("tcp", ":9990")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("listen", listener.Addr().String())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		conn,err := listener.Accept()
		if err != nil {
			panic(err)
		}
		//conn = UpdateConn(conn)
		go func(conn net.Conn) {
			defer wg.Done()
			time.Sleep(2 * time.Second)
			fmt.Println("read after 2s")
			defer conn.Close()
			var buf = make([]byte, 1024)
			conn.(*gramConn).conn.Read(make([]byte, 1))
			for {
				n,err := conn.Read(buf)
				if err != nil {
					fmt.Println("read error:", err.Error())
					break
				}
				fmt.Println("read len:", n, "read data:", string(buf[:n]))
			}
		}(conn)
	}()

	conn,err := Dial("tcp", ":9990")
	if err != nil {
		panic(err)
	}
	fmt.Println("dial:", conn.LocalAddr().String())

	//conn = UpdateConn(conn)

	for i := 0; i < 10; i++ {
		n,err := conn.Write([]byte(fmt.Sprintf("hello%03d", i)))
		if err != nil {
			panic(err)
		}
		fmt.Println(i, "send len:", n)
	}

	conn.Close()
	wg.Wait()
}
