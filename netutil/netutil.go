package netutil

import (
	"net"
	"time"
	"errors"
	"bytes"
	"fmt"
)

const MinRead = 512

const bufferSize = 1024 * 1024

var (
	ErrInvalidConn = errors.New("netutil: invalid conn")
)

type gramConn struct {
	conn   net.Conn
	head   header
	len    int
	rBuf   []byte
	wBuf   []byte
	hBuf   []byte
}

func (c *gramConn) Read(b []byte) (n int, err error) {
	if c.head.Length == 0 {
		var length = 0
		for {
			if n,err = c.conn.Read(c.hBuf[length:]); err != nil {
				return
			}
			length += n
			if length >= len(c.hBuf) {
				break
			}
		}
		fmt.Println("len:", length)
		fmt.Println("<read>:", string(c.hBuf))
		if err = c.head.decode(bytes.NewReader(c.hBuf)); err != nil {
			return
		}
		if err = c.head.verify(); err != nil {
			return
		}
	}
	if c.head.Length > 0 {
		if c.head.Length > uint64(len(c.rBuf)) {
			c.rBuf = make([]byte, c.head.Length * 2)
		}
		var length uint64 = 0
		for {
			n,err = c.conn.Read(c.rBuf[length:c.head.Length + length])
			if err != nil {
				return
			}
			length += uint64(n)
			if length >= c.head.Length {
				break
			}
		}
		n = int(length)
		copy(b, c.rBuf)
		c.head.Length = 0
		//if c.len == 0 {
		//	if n,err = c.conn.Read(c.rBuf[c.len:]); err != nil {
		//		return
		//	}
		//	c.len = n
		//}
		//if len(b) <= c.len {
		//	copy(b, c.rBuf)
		//	c.len = 0
		//} else {
		//	var length = 0
		//	length = copy(b, c.rBuf)
		//}
	}

	return
}

func (c *gramConn) Write(b []byte) (n int, err error) {
	var h = newHeader()
	if err = h.calc(b); err != nil {
		return
	}
	if err = h.verify(); err != nil {
		return
	}
	var buf = bytes.NewBuffer(c.wBuf)
	if err = h.encode(buf); err != nil {
		return
	}
	if n,err = buf.Write(b); err != nil {
		return
	}

	fmt.Println("<write>:", string(buf.Bytes()))

	n,err = c.conn.Write(buf.Bytes())
	if n >= headerSize() {
		n -= headerSize()
	}

	return
}

func (c *gramConn) Close() error {
	return c.conn.Close()
}

func (c *gramConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *gramConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *gramConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *gramConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *gramConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func updateConn(conn net.Conn, buf []byte) net.Conn {
	if _,ok := conn.(*gramConn); ok {
		return conn
	}

	if buf == nil || len(buf) == 0 {
		buf = make([]byte, bufferSize)
	}

	var c = new(gramConn)

	c.conn = conn
	c.rBuf = buf
	c.wBuf = make([]byte, bufferSize)
	c.len = 0
	c.hBuf = make([]byte, headerSize())

	return c
}

func readGram(conn net.Conn, capacity int64) (b []byte, err error) {
	if _,ok := conn.(*gramConn); !ok {
		return nil, ErrInvalidConn
	}

	return
}

func UpdateConn(conn net.Conn) net.Conn {
	return updateConn(conn, make([]byte, bufferSize))
}

func ReadGram(conn net.Conn) ([]byte, error) {
	return readGram(conn, MinRead)
}