package lodm

import (
	"encoding/binary"
	"image"
	"io"

	"github.com/flywave/go3d/mat3"
)

type TextureImage image.Image

type TextureData []byte

type Texture struct {
	Offset uint32
	Mat    mat3.T
}

func (m *Texture) address() int64 {
	return int64(m.Offset * LM_PADDING)
}

func (m *Texture) CalcSize() int64 {
	return int64(binary.Size(*m))
}

func (m *Texture) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *Texture) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
