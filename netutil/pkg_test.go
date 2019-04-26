package netutil

import (
	"testing"
	"fmt"
	"bytes"
)

func TestHeader(t *testing.T) {
	h := newHeader()
	fmt.Println("header size:", headerSize())
	fmt.Println("iHeader size:", iHeaderSize())
	var data = []byte("hello world")
	var err error
	if err = h.calc(data); err != nil {
		panic(err)
	}
	if err = h.verify(); err != nil {
		panic(err)
	}
	var buf = bytes.NewBuffer(nil)
	h.encode(buf)
	var h2 = newHeader()
	h2.decode(buf)
	if err = h2.verify(); err != nil {
		panic(err)
	}
}