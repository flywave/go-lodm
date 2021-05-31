package lodm

import (
	"encoding/binary"
	"io"

	"github.com/flywave/go3d/mat4"
)

type Instance struct {
	InstanceID []uint64
	Mat        []mat4.T
}

type InstanceNode struct {
	Node
	NInstance uint32
}

func (m *InstanceNode) CalcSize() int64 {
	return int64(binary.Size(Feature{}))
}

func (m *InstanceNode) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *InstanceNode) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
