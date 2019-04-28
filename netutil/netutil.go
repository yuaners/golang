package netutil

import (
	"net"
	"time"
	"errors"
	"io"
)

const bufferSize = 1024 * 1024

type Header interface {
	// Calculated data on the head
	Calc(data []byte) error
	// Header length
	Len() uint64
	// Encode header too writer
	Encode(writer io.Writer) error
	// Decode header from reader
	Decode(reader io.Reader) error
}

type packConn struct {
	// The original connection
	conn net.Conn
	// The DataGram head
	rHead Header
	wHead Header

	// read/write buffer
	rBuf []byte
	wBuf []byte
}

func (c *packConn) Read(b []byte) (n int, err error) {
	if c.rHead == nil {
		return c.conn.Read(b)
	}

	if err = c.rHead.Decode(c.conn); err != nil {
		return 0, err
	}

	if c.rHead.Len() == 0 {
		return 0, nil
	}

	if c.rHead.Len() > uint64(len(c.rBuf)) {
		defer func() {
			if recover() != nil {
				err = errors.New("make slice too large")
			}
		}()
		c.rBuf = make([]byte, c.rHead.Len() * 2)
	}

	var length uint64 = 0
	for {
		n,err = c.conn.Read(c.rBuf[length:length + c.rHead.Len()])
		if err != nil {
			return
		}
		length += uint64(n)
		if length >= c.rHead.Len() {
			break
		}
	}
	n = copy(b, c.rBuf[:length])

	return n, nil
}

func (c *packConn) Write(b []byte) (n int, err error) {
	if c.wHead == nil {
		return c.conn.Write(b)
	}

	if err = c.wHead.Calc(b); err != nil {
		return 0, err
	}

	if err = c.wHead.Encode(c.conn); err != nil {
		return 0, err
	}

	if n,err = c.conn.Write(b); err != nil {
		return 0, err
	}

	return
}

func (c *packConn) Close() error {
	return c.conn.Close()
}

func (c *packConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *packConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *packConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *packConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *packConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func updateConn(conn net.Conn, r, w Header, rBuf, wBuf []byte) net.Conn {
	if _,ok := conn.(*packConn); ok {
		return conn
	}

	if rBuf == nil || len(rBuf) == 0 {
		rBuf = make([]byte, bufferSize)
	}
	if wBuf == nil || len(wBuf) == 0 {
		wBuf = make([]byte, bufferSize)
	}

	var c = new(packConn)

	c.conn = conn
	c.rBuf = rBuf
	c.wBuf = wBuf
	c.rHead = r
	c.wHead = w

	return c
}

func UpdateConn(conn net.Conn) net.Conn {
	return updateConn(conn, newHeader(), newHeader(), make([]byte, bufferSize), make([]byte, bufferSize))
}

func Dial(network, address string) (net.Conn, error) {
	conn, err := net.Dial(network, address)
	if err == nil {
		conn = UpdateConn(conn)
	}

	return conn, err
}

type packListener struct {
	net.Listener
}

func (l *packListener) Accept() (net.Conn, error) {
	conn,err := l.Listener.Accept()
	if err == nil {
		conn = UpdateConn(conn)
	}

	return conn, err
}

func (l *packListener) Close() error {
	return l.Listener.Close()
}

func (l *packListener) Addr() net.Addr {
	return l.Listener.Addr()
}

func updateListener(listener net .Listener) net.Listener {
	if _,ok := listener.(*packListener); ok {
		return listener
	}
	var l = new(packListener)
	l.Listener = listener

	return l
}

func UpdateListener(listener net.Listener) net.Listener {
	return updateListener(listener)
}

func Listen(network, address string) (net.Listener, error) {
	listener,err := net.Listen(network, address)
	if err == nil {
		listener = updateListener(listener)
	}

	return listener, err
}