package lodm

import (
	"encoding/binary"
	"io"
)

type Patch struct {
	Node       uint32
	FaceOffset uint32
	TexID      uint32
	MtlID      uint32
	FeatID     uint32
}

func (m *Patch) SetNode(n uint32) {
	m.Node = n
}

func (m *Patch) GetNode() uint32 {
	return m.Node
}

func (m *Patch) CalcSize() int64 {
	return int64(binary.Size(*m))
}

func (m *Patch) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *Patch) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
