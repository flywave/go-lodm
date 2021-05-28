package lodm

import (
	"encoding/binary"

	"github.com/flywave/go3d/mat4"
)

const (
	MESH_PADDING = 256
)

var byteorder = binary.LittleEndian

type Storage struct {
	Header    Header
	Nodes     []Node
	Instances []InstanceNode
	Textures  []Texture
	Materials []Material
	Features  []Feature
	Sphere    Sphere
	Mat       mat4.T
}
