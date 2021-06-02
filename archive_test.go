package lodm

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestHeader(t *testing.T) {
	if si := binary.Size(Header{}); si != HeaderSize {
		t.FailNow()
	}

	var h Header

	h.Magic = MAGIC_BYTE

	var buf bytes.Buffer
	h.Write(&buf)

	b := buf.Bytes()
	if len(b) != HeaderSize {
		t.FailNow()
	}

	rbuf := bytes.NewBuffer(b)

	var h1 Header

	h1.Read(rbuf)

	if h1.Magic != MAGIC_BYTE {
		t.FailNow()
	}
}

func TestPadding(t *testing.T) {
	padding := calcPadding(213, 256)

	if padding != (256 - 213) {
		t.FailNow()
	}
}

func TestInstanceNodeId(t *testing.T) {

}
