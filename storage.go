package lodm

import (
	"encoding/binary"

	"github.com/flywave/go-corto"
	"github.com/flywave/go-draco"

	"github.com/flywave/go3d/mat4"
	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
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
	CoordQ     int
	CoordBits  int
	NormalBits int
	UvBits     int
	ColorBits  int
}

func Compress(header Header, node *Node, mesh *NodeMesh, patches []Patch, setting CompressSetting) []byte {
	sig := header.Sign
	if (sig.Flags & CORTO) > 0 {
		ctx := corto.NewEncoderContext(setting.CoordQ)
		if setting.CoordBits > 0 {
			ctx.VertexBits = setting.CoordQ
		}
		ctx.NormBits = setting.NormalBits
		ctx.UvBits = float32(setting.UvBits) / 512
		ctx.ColorBits = [4]int{setting.ColorBits, setting.ColorBits, setting.ColorBits, setting.ColorBits}

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
		enc := draco.NewEncoder()

		if setting.CoordBits > 0 {
			enc.SetAttributeQuantization(draco.GAT_POSITION, int32(setting.CoordBits))
		}
		if setting.NormalBits > 0 {
			enc.SetAttributeQuantization(draco.GAT_NORMAL, int32(setting.NormalBits))
		}
		if setting.UvBits > 0 {
			enc.SetAttributeQuantization(draco.GAT_TEX_COORD, int32(setting.UvBits))
		}
		if setting.ColorBits > 0 {
			enc.SetAttributeQuantization(draco.GAT_COLOR, int32(setting.ColorBits))
		}

		if node.NFace == 0 {
			builder := draco.NewPointCloudBuilder()
			builder.Start(int(node.NVert))

			builder.SetAttribute(int(node.NVert), mesh.Verts[:], draco.GAT_POSITION)

			if sig.Vertex.HasNormals() {
				builder.SetAttribute(int(node.NVert), mesh.Normals[:], draco.GAT_NORMAL)
			}

			if sig.Vertex.HasColors() {
				builder.SetAttribute(int(node.NVert), mesh.Colors[:], draco.GAT_COLOR)
			}
			pc := builder.GetPointCloud()
			_, buf := enc.EncodePointCloud(pc)
			return buf
		} else {
			builder := draco.NewMeshBuilder()
			size := int(node.NFace)

			builder.Start(size)

			face_points := make([]vec3.T, size*3)

			var face_normals [][3]int16
			if sig.Vertex.HasNormals() {
				face_normals = make([][3]int16, size*3)
			}
			var face_colors [][4]byte
			if sig.Vertex.HasColors() {
				face_colors = make([][4]byte, size*3)
			}
			var face_texcoords []vec2.T
			if sig.Vertex.HasTextures() {
				face_texcoords = make([]vec2.T, size*3)
			}

			for i := range mesh.Faces {
				face_points[i*3] = mesh.Verts[int(mesh.Faces[i][0])]
				face_points[i*3+1] = mesh.Verts[int(mesh.Faces[i][1])]
				face_points[i*3+2] = mesh.Verts[int(mesh.Faces[i][2])]
				if sig.Vertex.HasNormals() {
					face_normals[i*3] = mesh.Normals[int(mesh.Faces[i][0])]
					face_normals[i*3+1] = mesh.Normals[int(mesh.Faces[i][1])]
					face_normals[i*3+2] = mesh.Normals[int(mesh.Faces[i][2])]
				}
				if sig.Vertex.HasColors() {
					face_colors[i*3] = mesh.Colors[int(mesh.Faces[i][0])]
					face_colors[i*3+1] = mesh.Colors[int(mesh.Faces[i][1])]
					face_colors[i*3+2] = mesh.Colors[int(mesh.Faces[i][2])]
				}
				if sig.Vertex.HasTextures() {
					face_texcoords[i*3] = mesh.Texcoords[int(mesh.Faces[i][0])]
					face_texcoords[i*3+1] = mesh.Texcoords[int(mesh.Faces[i][1])]
					face_texcoords[i*3+2] = mesh.Texcoords[int(mesh.Faces[i][2])]
				}
			}

			builder.SetAttribute(size, face_points[:], draco.GAT_POSITION)
			if sig.Vertex.HasNormals() {
				builder.SetAttribute(size, face_normals[:], draco.GAT_NORMAL)
			}
			if sig.Vertex.HasColors() {
				builder.SetAttribute(size, face_colors[:], draco.GAT_COLOR)
			}
			if sig.Vertex.HasTextures() {
				builder.SetAttribute(size, face_texcoords[:], draco.GAT_COLOR)
			}
			mesh := builder.GetMesh()
			_, buf := enc.EncodeMesh(mesh)
			return buf
		}
	}
	return nil
}
