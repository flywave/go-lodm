package lodm

import (
	"encoding/binary"
	"io"

	"github.com/flywave/go3d/mat4"
)

type Instance struct {
	Node        uint32
	InstanceID  uint32
	InstanceMat mat4.T
}

func (m *Instance) CalcSize() int64 {
	return int64(binary.Size(*m))
}

func (m *Instance) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *Instance) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
