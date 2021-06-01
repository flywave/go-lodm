package lodm

import (
	"encoding/binary"
	"io"
	"reflect"
	"unsafe"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

type NodeData []byte

type NodeMesh struct {
	Verts     []vec3.T
	Faces     [][3]uint16
	Normals   [][3]int16
	Texcoords []vec2.T
	Colors    [][4]byte
}

func (m *NodeMesh) Empty() bool {
	return len(m.Verts) == 0
}

func (m *NodeMesh) CalcSize() int64 {
	return int64((len(m.Verts) * 3 * 4) + (len(m.Faces) * 3 * 2) + (len(m.Texcoords) * 2 * 4) + (len(m.Normals) * 3 * 2) + (len(m.Colors) * 4))
}

func (m *NodeMesh) Read(reader io.Reader, node *Node, header *Header) error {
	sig := header.Sign

	m.Verts = make([]vec3.T, node.NVert)

	var vertsSlice []float32
	vertsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&vertsSlice)))
	vertsHeader.Cap = int(node.NVert * 3)
	vertsHeader.Len = int(node.NVert * 3)
	vertsHeader.Data = uintptr(unsafe.Pointer(&m.Verts[0]))

	if err := binary.Read(reader, byteorder, vertsSlice); err != nil {
		return err
	}

	if node.NFace != 0 {
		m.Faces = make([][3]uint16, node.NFace)

		var facesSlice []uint16
		facesHeader := (*reflect.SliceHeader)((unsafe.Pointer(&facesSlice)))
		facesHeader.Cap = int(node.NFace * 3)
		facesHeader.Len = int(node.NFace * 3)
		facesHeader.Data = uintptr(unsafe.Pointer(&m.Faces[0]))

		if err := binary.Read(reader, byteorder, facesSlice); err != nil {
			return err
		}
	}

	if sig.Vertex.HasNormals() {
		m.Normals = make([][3]int16, node.NVert)

		var normalsSlice []int16
		normalsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&normalsSlice)))
		normalsHeader.Cap = int(node.NVert * 3)
		normalsHeader.Len = int(node.NVert * 3)
		normalsHeader.Data = uintptr(unsafe.Pointer(&m.Normals[0]))

		if err := binary.Read(reader, byteorder, normalsSlice); err != nil {
			return err
		}
	}

	if sig.Vertex.HasTextures() {
		m.Texcoords = make([]vec2.T, node.NVert)

		var texcoordsSlice []float32
		texcoordsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&texcoordsSlice)))
		texcoordsHeader.Cap = int(node.NVert * 2)
		texcoordsHeader.Len = int(node.NVert * 2)
		texcoordsHeader.Data = uintptr(unsafe.Pointer(&m.Texcoords[0]))

		if err := binary.Read(reader, byteorder, texcoordsSlice); err != nil {
			return err
		}
	}

	if sig.Vertex.HasColors() {
		m.Colors = make([][4]byte, node.NVert)

		var colorsSlice []byte
		colorsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&colorsSlice)))
		colorsHeader.Cap = int(node.NVert * 4)
		colorsHeader.Len = int(node.NVert * 4)
		colorsHeader.Data = uintptr(unsafe.Pointer(&m.Colors[0]))

		if err := binary.Read(reader, byteorder, colorsSlice); err != nil {
			return err
		}
	}

	return nil
}

func (m *NodeMesh) Write(writer io.Writer, node *Node, header *Header) error {
	sig := header.Sign

	node.NVert = uint16(len(m.Verts))

	var vertsSlice []float32
	vertsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&vertsSlice)))
	vertsHeader.Cap = int(node.NVert * 3)
	vertsHeader.Len = int(node.NVert * 3)
	vertsHeader.Data = uintptr(unsafe.Pointer(&m.Verts[0]))

	if err := binary.Write(writer, byteorder, vertsSlice); err != nil {
		return err
	}

	if f := len(m.Faces); f > 0 {
		node.NFace = uint16(f)

		var facesSlice []uint16
		facesHeader := (*reflect.SliceHeader)((unsafe.Pointer(&facesSlice)))
		facesHeader.Cap = int(node.NFace * 3)
		facesHeader.Len = int(node.NFace * 3)
		facesHeader.Data = uintptr(unsafe.Pointer(&m.Faces[0]))

		if err := binary.Write(writer, byteorder, facesSlice); err != nil {
			return err
		}
	}

	if sig.Vertex.HasNormals() && m.HasNormal() {
		var normalsSlice []int16
		normalsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&normalsSlice)))
		normalsHeader.Cap = int(node.NVert * 3)
		normalsHeader.Len = int(node.NVert * 3)
		normalsHeader.Data = uintptr(unsafe.Pointer(&m.Normals[0]))

		if err := binary.Write(writer, byteorder, normalsSlice); err != nil {
			return err
		}
	}

	if sig.Vertex.HasTextures() && m.HasTexcoord() {
		var texcoordsSlice []float32
		texcoordsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&texcoordsSlice)))
		texcoordsHeader.Cap = int(node.NVert * 2)
		texcoordsHeader.Len = int(node.NVert * 2)
		texcoordsHeader.Data = uintptr(unsafe.Pointer(&m.Texcoords[0]))

		if err := binary.Write(writer, byteorder, texcoordsSlice); err != nil {
			return err
		}
	}

	if sig.Vertex.HasColors() && m.HasColor() {
		var colorsSlice []byte
		colorsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&colorsSlice)))
		colorsHeader.Cap = int(node.NVert * 4)
		colorsHeader.Len = int(node.NVert * 4)
		colorsHeader.Data = uintptr(unsafe.Pointer(&m.Colors[0]))

		if err := binary.Write(writer, byteorder, colorsSlice); err != nil {
			return err
		}
	}
	return nil
}

func (m *NodeMesh) HasFace() bool {
	return len(m.Faces) > 0
}

func (m *NodeMesh) HasTexcoord() bool {
	return len(m.Texcoords) > 0
}

func (m *NodeMesh) HasNormal() bool {
	return len(m.Normals) > 0
}

func (m *NodeMesh) HasColor() bool {
	return len(m.Colors) > 0
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
