package lodm

import (
	"encoding/binary"
	"io"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

type NodeMesh struct {
	Verts     []vec3.T
	Faces     [][3]uint16
	Texcoords []vec2.T
	Normals   [][3]int16
	Colors    [][4]byte
}

type Node struct {
	Offset      uint32
	NVert       uint16
	NFace       uint16
	Error       float32
	Cone        Cone3s
	Sphere      Sphere
	TightRadius float32
	FirstPatch  uint32
}

func (m *Node) TightSphere() Sphere {
	return Sphere{m.Sphere[0], m.Sphere[1], m.Sphere[2], m.TightRadius}
}

func (m *Node) CalcSize() int64 {
	return int64(binary.Size(Feature{}))
}

func (m *Node) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *Node) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
