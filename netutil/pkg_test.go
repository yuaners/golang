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
	fmt.Println("calc finish.")
	if err = h.verify(); err != nil {
		panic(err)
	}
	fmt.Println("verify finish.")
	var buf = bytes.NewBuffer(nil)
	h.encode(buf)
	fmt.Println("encode finish.")
	var h2 = newHeader()
	h2.decode(buf)
	fmt.Println("decode finish.")
	if err = h2.verify(); err != nil {
		panic(err)
	}
	fmt.Println("verify finish.")
	h2.reset()
	fmt.Println("reset finish.")
	if err = h2.verify(); err != nil {
		fmt.Println("success reset:", err.Error())
	}
	fmt.Println("verify finish.")
}