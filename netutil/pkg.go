package netutil

import (
	"io"
	"fmt"
	"time"
	"bytes"
	"unsafe"
	"crypto/md5"
	"encoding/binary"
)

const protocolMagic = 0x5F6C656E6F766F5F

const protocolVersion = 1 << 24

type iHeader struct {
	Magic     uint64
	Version   uint64
	Length    uint64
}

type header struct {
	iHeader
	Timestamp int64
	Md5       [md5.Size]byte
}

func (h *header) calc(b []byte) error {
	h.Magic = protocolMagic
	h.Version = protocolVersion
	h.Length = uint64(len(b))
	if h.Length > 0 {
		var buffer = bytes.NewBuffer(make([]byte, iHeaderSize()))
		if err := binary.Write(buffer, binary.BigEndian, h.iHeader); err != nil {
			return err
		}
		var m = md5.Sum(buffer.Bytes())
		copy(h.Md5[:], m[:])
	}
	h.Timestamp = time.Now().UnixNano()

	return nil
}

func (h *header) verify() error {
	if h.Magic != protocolMagic {
		return fmt.Errorf("invalid magic: <%02X>", h.Magic)
	}
	if h.Version != protocolVersion {
		return fmt.Errorf("invalid version: <%02X>", h.Version)
	}
	if h.Length < 0 {
		return fmt.Errorf("invalid length: <%d>", h.Length)
	}
	if h.Length > 0 {
		var buffer = bytes.NewBuffer(make([]byte, iHeaderSize()))
		if err := binary.Write(buffer, binary.BigEndian, h.iHeader); err != nil {
			return err
		}
		var m = md5.Sum(buffer.Bytes())
		for i,v := range m {
			if v != h.Md5[i] {
				return fmt.Errorf("invalid md5: <%02x>", h.Md5)
			}
		}
	}

	return nil
}

func (h *header) encode(writer io.Writer) error {
	return binary.Write(writer, binary.BigEndian, h)
}

func (h *header) decode(reader io.Reader) error {
	return binary.Read(reader, binary.BigEndian, h)
}

func headerSize() int {
	return int(unsafe.Sizeof(header{}))
}

func iHeaderSize() int {
	return int(unsafe.Sizeof(header{})) - md5.Size - 8
}

func newHeader() *header {
	return new(header)
}