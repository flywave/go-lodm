package lodm

import (
	"encoding/binary"
	"io"

	"github.com/flywave/go3d/vec3"
)

type FeatureData string

type Feature struct {
	Offset uint32
	Type   uint32
	ID     uint32
	Sphere [4]float32
	Box    vec3.Box
}

func (m *Feature) CalcSize() int64 {
	return int64(binary.Size(Feature{}))
}

func (m *Feature) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *Feature) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}