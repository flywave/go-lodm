package lodm

import (
	"encoding/binary"
	"io"

	"github.com/flywave/go3d/vec3"
)

type AttributeType uint8
type ComponentType uint8

const (
	ATTR_NONE           AttributeType = 0
	ATTR_BYTE           AttributeType = 1
	ATTR_UNSIGNED_BYTE  AttributeType = 2
	ATTR_SHORT          AttributeType = 3
	ATTR_UNSIGNED_SHORT AttributeType = 4
	ATTR_INT            AttributeType = 5
	ATTR_UNSIGNED_INT   AttributeType = 6
	ATTR_FLOAT          AttributeType = 7
	ATTR_DOUBLE         AttributeType = 8
)

var (
	typeSize = []int{0, 1, 1, 2, 2, 4, 4, 4, 8}
)

type Attribute struct {
	Type   byte
	Number byte
}

func (a Attribute) Size() int {
	return int(a.Number) * typeSize[a.Type]
}

func (a Attribute) Valid() bool { return a.Type != 0 }

type Element struct {
	Attributes [8]Attribute
}

func (e *Element) SetComponent(c int, a Attribute) {
	if c < 8 {
		e.Attributes[c] = a
	} else {
		panic("error")
	}
}

func (e *Element) Size() int {
	s := 0
	for i := 0; i < 8; i++ {
		s += e.Attributes[i].Size()
	}
	return s
}

const (
	VERTEX_COORD ComponentType = 0
	VERTEX_NORM  ComponentType = 1
	VERTEX_COLOR ComponentType = 2
	VERTEX_TEX   ComponentType = 3
	VERTEX_DATA0 ComponentType = 4
)

type VertexElement struct {
	Element
}

func (e *VertexElement) HasNormals() bool {
	return e.Attributes[VERTEX_NORM].Valid()
}

func (e *VertexElement) HasColors() bool {
	return e.Attributes[VERTEX_COLOR].Valid()
}

func (e *VertexElement) HasTextures() bool {
	return e.Attributes[VERTEX_TEX].Valid()
}

func (e *VertexElement) HasData(i int) bool {
	return e.Attributes[int(VERTEX_DATA0)+i].Valid()
}

const (
	FACE_INDEX ComponentType = 0
	FACE_NORM  ComponentType = 1
	FACE_COLOR ComponentType = 2
	FACE_TEX   ComponentType = 3
	FACE_DATA0 ComponentType = 4
)

type FaceElement struct {
	Element
}

func (e *FaceElement) HasIndex() bool {
	return e.Attributes[FACE_INDEX].Valid()
}

func (e *FaceElement) HasNormals() bool {
	return e.Attributes[FACE_NORM].Valid()
}

func (e *FaceElement) HasColors() bool {
	return e.Attributes[FACE_COLOR].Valid()
}

func (e *FaceElement) HasTextures() bool {
	return e.Attributes[FACE_TEX].Valid()
}

func (e *FaceElement) HasData(i int) bool {
	return e.Attributes[int(FACE_DATA0)+i].Valid()
}

type FlagType uint32

const (
	PTJPG FlagType = 0x1
	PTPNG FlagType = 0x2
	CORTO FlagType = 0x4
	DRACO FlagType = 0x8
	TILE  FlagType = 0x16
)

type Signature struct {
	Vertex VertexElement
	Face   FaceElement
	Flags  FlagType
}

func (s *Signature) SetFlag(f FlagType) {
	s.Flags |= f
}
func (s *Signature) UnsetFlag(f FlagType) {
	s.Flags &= ^f
}

func (s *Signature) HasPTextures() bool {
	return ((s.Flags & (PTJPG | PTPNG)) > 0)
}

func (s *Signature) IsCompressed() bool {
	return (s.Flags & (CORTO | DRACO)) > 0
}

func (s *Signature) IsTile() bool {
	return ((s.Flags | TILE) > 0)
}

const (
	CurrentVersion = 0
)

var (
	MAGIC_STR = []byte("fwlm")
)

type Header struct {
	Magic      uint32
	Version    uint32
	NVert      uint64
	NFace      uint64
	Sign       Signature
	NNodes     uint32
	NInstances uint32
	NPatches   uint32
	NTextures  uint32
	NMaterials uint32
	NFeatures  uint32
	Sphere     Sphere
	Matrix     vec3.T
	Tile       [3]uint32
	Padding    [80]byte
}

func (m *Header) CalcInstanceId(node uint32) uint32 {
	return node - m.NNodes
}

func (m *Header) CalcNodeId(insId uint32) uint32 {
	return insId + m.NNodes
}

func (m *Header) CalcSize() int64 {
	return int64(binary.Size(*m))
}

func (m *Header) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *Header) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
