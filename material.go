package lodm

import (
	"encoding/binary"
	"io"
)

const (
	MTL_BASE    = 0
	MTL_LAMBERT = 1
	MTL_PHONG   = 2
	MTL_PBR     = 3
)

const (
	COLOR   = 0x1
	TEXTURE = 0x2
	BUMP    = 0x4
)

type Material struct {
	Type               uint32
	Mode               uint32
	Color              [3]byte
	Ambient            [3]byte
	Emissive           [3]byte
	Specular           [3]byte
	Opacity            float32
	Shininess          float32
	Metallic           float32
	Roughness          float32
	Reflectance        float32
	ClearcoatThickness float32
	ClearcoatRoughness float32
	Anisotropy         float32
	AnisotropyRotation float32
}

func (m *Material) CalcSize() int64 {
	return int64(binary.Size(*m))
}

func (m *Material) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *Material) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
