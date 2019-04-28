package netutil

import (
	"io"
		"time"
	"bytes"
	"unsafe"
	"crypto/md5"
	"encoding/binary"
		"fmt"
	"sync"
)

const protocolMagic = 0x5F6C656E6F766F5F

const protocolVersion = 1 << 24


const (
	errInvalidMagic = iota + 1
	errInvalidVersion
	errInvalidLength
	errInvalidMd5
)

type headerError struct {
	code int
	err  error
}

func (h *headerError) Error() string {
	return h.err.Error()
}

func newHeaderError(code int, err error) error {
	var e = new(headerError)
	e.code = code
	e.err = err

	return e
}

type iHeader struct {
	Magic     uint64
	Version   uint64
	Length    uint64
	Timestamp int64
}

type header struct {
	iHeader
	Md5       [md5.Size]byte
}

func (h *header) Calc(b []byte) error {
	h.Magic = protocolMagic
	h.Version = protocolVersion
	h.Length = uint64(len(b))
	h.Timestamp = time.Now().UnixNano()
	if h.Length > 0 {
		var buffer = bytes.NewBuffer(make([]byte, iHeaderSize()))
		if err := binary.Write(buffer, binary.BigEndian, h.iHeader); err != nil {
			return err
		}
		h.Md5 = md5.Sum(buffer.Bytes())
	}

	return nil
}

func (h *header) Len() uint64 {
	return h.Length
}

func (h *header) Encode(writer io.Writer) error {
	return binary.Write(writer, binary.BigEndian, h)
}

func (h *header) Decode(reader io.Reader) error {
	var err error
	var buf = pool.Get().([]byte)
	defer pool.Put(buf)
	h.reset()
	if _,err = io.ReadFull(reader, buf[:headerSize()]); err != nil {
		return err
	}

	for {
		for i := 0; i <= len(buf) - headerSize(); i++ {
			if err = binary.Read(bytes.NewReader(buf[i:]), binary.BigEndian, h); err != nil {
				return err
			}
			if err = h.verify(); err == nil {
				return nil
			}
		}
		buf = append(buf, 0)
		if _,err = io.ReadFull(reader, buf[len(buf) - 1:]); err != nil {
			return err
		}
	}

	return nil
}

func (h *header) verify() error {
	if h.Magic != protocolMagic {
		return newHeaderError(errInvalidMagic, fmt.Errorf("invalid magic: <%02X>", h.Magic))
	}

	if h.Version != protocolVersion {
		return newHeaderError(errInvalidVersion, fmt.Errorf("invalid version: <%02X>", h.Version))
	}
	if h.Length < 0 {
		return newHeaderError(errInvalidLength, fmt.Errorf("invalid length: <%d>", h.Length))
	}
	if h.Length > 0 {
		var buffer = bytes.NewBuffer(make([]byte, iHeaderSize()))
		if err := binary.Write(buffer, binary.BigEndian, h.iHeader); err != nil {
			return err
		}
		var m = md5.Sum(buffer.Bytes())
		for i,v := range m {
			if v != h.Md5[i] {
				return newHeaderError(errInvalidMd5, fmt.Errorf("invalid md5: <%02x>", h.Md5))
			}
		}
	}

	return nil
}

func (h *header) reset() {
	h.Magic = 0
	h.Version = 0
	h.Length = 0
	h.Timestamp = 0
	for i,_ := range h.Md5 {
		h.Md5[i] = 0
	}
}

func headerSize() int {
	return int(unsafe.Sizeof(header{}))
}

func iHeaderSize() int {
	return int(unsafe.Sizeof(header{}.iHeader))
}

func newHeader() *header {
	return new(header)
}

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, headerSize(), headerSize() * 100)
	},
}