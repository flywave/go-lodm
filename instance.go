package lodm

import (
	"encoding/binary"
	"io"

	"github.com/flywave/go3d/quaternion"
	"github.com/flywave/go3d/vec3"
)

type Instance struct {
	ID       uint32
	Position vec3.T
	Scale    [3]float32
	Rotate   quaternion.T
}

type InstanceNode struct {
	Node
	EastNorthUp uint16
	NInstance   uint16
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
