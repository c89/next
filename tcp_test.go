package next

import (
	"bytes"
	_ "encoding/binary"
	"fmt"
	"testing"
)

func TestTcpPack(t *testing.T) {
	tcp := NewTcp()

	s := "GOOD, 你好"

	w := bytes.NewBuffer([]byte{})
	err := tcp.Pack(w, []byte(s))
	if err != nil {
		t.Error(err)
	}

	pack := w.Bytes()
	fmt.Printf("%v\n", pack)

	buf := bytes.NewReader(pack)
	data, err := tcp.Unpack(buf)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", string(data))

	t.Logf("sussess")
}

func TestTcpUnpack(t *testing.T) {
	b := []byte{0xAA, 0x00, 0x00, 0x00, 0x01, 0x11, 0x55}
	//buf := bytes.NewReader(b) // 16KB
	//	buf := new(bytes.Buffer)
	//	binary.Write(buf, binary.BigEndian, b)
	buf := bytes.NewReader(b)

	tcp := NewTcp()
	_, err := tcp.Unpack(buf)
	if err != nil {
		t.Error(err)
	}

	t.Logf("sussess")
}
