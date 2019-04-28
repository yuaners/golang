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
	if err = h.Calc(data); err != nil {
		panic(err)
	}
	fmt.Println("calc finish.")
	if err = h.verify(); err != nil {
		panic(err)
	}
	fmt.Println("verify finish.")
	var buf = bytes.NewBuffer(nil)
	h.Encode(buf)
	fmt.Println("encode finish.")
	var h2 = newHeader()
	h2.Decode(buf)
	fmt.Println("decode finish.")
	if err = h2.verify(); err != nil {
		panic(err)
	}
	fmt.Println("h2 len:", h2.Len())
	fmt.Println("verify finish.")
	h2.reset()
	fmt.Println("reset finish.")
	if err = h2.verify(); err != nil {
		fmt.Println("success reset:", err.Error())
	}
	fmt.Println("verify finish.")
}

func TestHeader_Calc(t *testing.T) {
	tests := map[string]struct {
		input []byte
		wantLen int
	} {
		"simple1": {input:[]byte("hello"), wantLen:len("hello")},
		"simple2": {input:[]byte("hello "), wantLen:len("hello ")},
		"simple3": {input:[]byte(" h e l l o "), wantLen:len(" h e l l o ")},
		"simple4": {input:[]byte("北京"), wantLen:len("北京")},
		"simple5": {input:[]byte("@。#"), wantLen:len("@。#")},
	}

	for name,tc := range tests {
		t.Run(name, func(t *testing.T) {
			h := newHeader()
			if err := h.Calc(tc.input); err != nil {
				t.Error(err)
			}
			if h.Len() != uint64(tc.wantLen) {
				t.Error("header len", h.Len(), "!=", tc.wantLen)
			}
			if err := h.verify(); err != nil {
				t.Error(err)
			}
		})
	}
}