package netutil

import (
	"net"
	"time"
	"errors"
	"bytes"
	"encoding/binary"
	"io"
	)

const MinRead = 512

const bufferSize = 1024 * 1024

var (
	ErrInvalidConn = errors.New("net util: invalid conn")

	ErrUnexpectedLength = errors.New("net util: unexpected length")
)

type gramConn struct {
	conn   net.Conn
	head   header
	len    int
	rBuf   []byte
	wBuf   []byte
	hBuf   []byte
}

func (c *gramConn) readFull(b []byte) error {
	return readFull(c.conn, b)
}

func (c *gramConn) readHeader() error {
	var length = 0
	var n int
	var err error
	for {
		if n,err = c.conn.Read(c.hBuf[length:]); err != nil {
			return err
		}
		length += n
		if length >= len(c.hBuf) {
			break
		}
	}

	return nil
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
		if err = c.head.decode(bytes.NewReader(c.hBuf)); err != nil {
			return
		}
		if err = c.head.verify(); err != nil {
			if err != ErrInvalidMagic {
				return
			}
			RESET:
			if err == ErrInvalidMagic {
				var buf = make([]byte, 0, 4 * 4096)
				buf = append(buf, c.hBuf...)
				for {
					if err = c.readHeader(); err != nil {
						return
					}
					buf = append(buf, c.hBuf...)
					var index = -1
					var w = bytes.NewBuffer(nil)
					if err = binary.Write(w, binary.BigEndian, uint64(protocolMagic)); err != nil {
						return
					}
					if index = bytes.Index(buf, w.Bytes()); index != -1 {
						if len(buf) - index < headerSize() {
							var diff = headerSize() - len(buf) + index
							var diffBuf = make([]byte, diff)
							if err = readFull(c.conn, diffBuf); err != nil {
								return
							}
							buf = append(buf, diffBuf...)
						}
						if err = c.head.decode(bytes.NewReader(buf[index:index + headerSize()])); err != nil {
							return
						}
						if err = c.head.verify(); err != nil {
							goto RESET
						}
						break
					}
				}
			}
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
		n = copy(b, c.rBuf[:length])
		c.head.Length = 0
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
	buf.Reset()
	if err = h.encode(buf); err != nil {
		return
	}
	if n,err = buf.Write(b); err != nil {
		return
	}

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

func Dial(network, address string) (net.Conn, error) {
	conn, err := net.Dial(network, address)
	if err == nil {
		conn = UpdateConn(conn)
	}

	return conn, err
}

type gramListener struct {
	net.Listener
}

func (l *gramListener) Accept() (net.Conn, error) {
	conn,err := l.Listener.Accept()
	if err == nil {
		conn = UpdateConn(conn)
	}

	return conn, err
}

func (l *gramListener) Close() error {
	return l.Listener.Close()
}

func (l *gramListener) Addr() net.Addr {
	return l.Listener.Addr()
}

func updateListener(listener net .Listener) net.Listener {
	if _,ok := listener.(*gramListener); ok {
		return listener
	}
	var l = new(gramListener)
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


func readFull(reader io.Reader, buf []byte) error {
	if buf == nil || len(buf) == 0 {
		return nil
	}
	var n int
	var err error
	var length int
	for {
		n,err = reader.Read(buf[n:])
		if err != nil {
			return err
		}
		length += n
		if length >= len(buf) {
			break
		}
	}
	if length != len(buf) {
		return ErrUnexpectedLength
	}

	return nil
}

// ReadFull read full buf from reader.
func ReadFull(reader io.Reader, buf []byte) error {
	return readFull(reader, buf)
}