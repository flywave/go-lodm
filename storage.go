package lodm

import (
	"encoding/binary"

	"github.com/flywave/go-corto"
	"github.com/flywave/go-draco"

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

type CompressSetting struct {
	CoordQ int
}

func Compress(header Header, node *Node, mesh *NodeMesh, patches []Patch, setting CompressSetting) []byte {
	sig := header.Sign
	if (sig.Flags & CORTO) > 0 {
		ctx := corto.NewEncoderContext(setting.CoordQ)
		geom := &corto.Geom{}

		for p := 0; p < len(patches); p++ {
			geom.Groups = append(geom.Groups, int(patches[p].Offset))
		}

		geom.Vertices = mesh.Verts[:]

		if node.NFace == 0 {
			geom.Indices16 = make([]corto.Face16, node.NFace)
			for i := 0; i < int(node.NFace); i++ {
				geom.Indices16[i] = corto.Face16{mesh.Faces[i][0], mesh.Faces[i][1], mesh.Faces[i][2]}
			}
		}

		if sig.Vertex.HasNormals() {
			geom.Normals16 = make([]corto.Normal16, node.NVert)
			for i := 0; i < int(node.NVert); i++ {
				geom.Normals16[i] = corto.Normal16{mesh.Normals[i][0], mesh.Normals[i][1], mesh.Normals[i][2]}
			}
		}

		if sig.Vertex.HasColors() {
			geom.Colors = make([]corto.Color, node.NVert)
			for i := 0; i < int(node.NVert); i++ {
				geom.Colors[i] = corto.Color{mesh.Colors[i][0], mesh.Colors[i][1], mesh.Colors[i][2], mesh.Colors[i][3]}
			}
		}

		if sig.Vertex.HasTextures() {
			geom.TexCoord = mesh.Texcoords[:]
		}

		return corto.EncodeGeom(ctx, geom)
	} else if (sig.Flags & DRACO) > 0 {
		dmesh := draco.NewMesh()
		d := NewDecoder()
		err := d.DecodeMesh(m, []byte{1, 2, 3})
	}
	return nil
}
