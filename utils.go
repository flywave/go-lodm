package lodm

import (
	"bytes"
	"image/jpeg"
	"image/png"
	"reflect"
	"unsafe"

	"github.com/flywave/go-corto"
	"github.com/flywave/go-draco"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

type Cone3s [4]int16

type Sphere [4]float32

func (s Sphere) Radius() float32 {
	return s[3]
}

func (s Sphere) Center() vec3.T {
	return vec3.T{s[0], s[1], s[2]}
}

func (s Sphere) IsEmpty() bool {
	return s[3] == 0
}

func (s Sphere) Add(sphere Sphere) {
	if s.IsEmpty() {
		s = sphere
		return
	}
	c1 := sphere.Center()
	c2 := s.Center()
	dist := vec3.Sub(&c1, &c2)
	distance := dist.Length()
	fartest := sphere.Radius() + distance
	if fartest <= s.Radius() {
		return
	}

	nearest := sphere.Radius() - distance
	if nearest >= s.Radius() {
		s = sphere
		return
	}

	if distance < 0.001*(s.Radius()+sphere.Radius()) {
		s[3] += distance
		return
	}
	delta := ((fartest - s.Radius()) / (distance * 2))
	t := vec3.T{delta, delta, delta}
	dist.Mul(&t)
	s[0] += dist[0]
	s[1] += dist[1]
	s[2] += dist[2]
	s[3] = (s.Radius() + fartest) / 2
}

func (s Sphere) Dist(p Sphere) float32 {
	pc := p.Center()
	sc := s.Center()
	dist := vec3.Sub(&pc, &sc)
	return dist.Length()
}

func (s Sphere) IsIn(p Sphere) bool {
	pc := p.Center()
	sc := s.Center()
	dist := vec3.Sub(&pc, &sc)
	distance := dist.Length()
	return distance+p.Radius() < s.Radius()
}

func calcPadding(offset, paddingUnit uint32) uint32 {
	padding := offset % paddingUnit
	if padding != 0 {
		padding = paddingUnit - padding
	}
	return padding
}

const (
	LM_JPEG_QUALITY int = 70
)

type CompressSetting struct {
	CoordQ     float32
	CoordBits  int
	NormalBits int
	UvBits     int
	ColorBits  int
	AlphaBits  int
}

var (
	DEFAULE_COMPRESS_SETTING = CompressSetting{CoordQ: 0, CoordBits: 1, NormalBits: 10, ColorBits: 6, AlphaBits: 5, UvBits: 1}
)

func CompressNode(header Header, node *Node, mesh *NodeMesh, patches []Patch, setting *CompressSetting) NodeData {
	buf := compressNodeMesh(header, node, mesh, patches, setting)
	padding := calcPadding(uint32(len(buf)), LM_PADDING)
	if padding == 0 {
		return buf
	}
	for i := 0; i < int(padding); i++ {
		buf = append(buf, byte(0))
	}
	return buf
}

func decompressNodeMesh(buf []byte, header Header, node *Node, mesh *NodeMesh) error {
	sign := &header.Sign
	if (sign.Flags & CORTO) > 0 {
		ctx := &corto.DecoderContext{}
		ctx.NFace = uint32(node.NFace)
		ctx.NVert = uint32(node.NVert)
		ctx.ColorsComponents = 4
		ctx.Index16 = true
		ctx.Normal16 = true
		geom := corto.DecodeGeom(ctx, buf)
		mesh.Verts = geom.Vertices[:]

		mesh.Normals = make([][3]int16, len(geom.Normals16))
		for i := range geom.Normals16 {
			mesh.Normals[i] = [3]int16(geom.Normals16[i])
		}
		if len(geom.TexCoord) > 0 {
			mesh.Texcoords = geom.TexCoord[:]
		}
		if len(geom.Indices16) > 0 {
			mesh.Faces = make([][3]uint16, len(geom.Indices16))
			for i := range geom.Indices16 {
				mesh.Faces[i] = [3]uint16(geom.Indices16[i])
			}
		}
		if len(geom.Colors) > 0 {
			mesh.Colors = make([][4]byte, len(geom.Colors))
			for i := range geom.Colors {
				mesh.Colors[i] = [4]byte(geom.Colors[i])
			}
		}
	} else {
		if node.NFace == 0 {
			m := draco.NewPointCloud()
			denc := draco.NewDecoder()
			err := denc.DecodePointCloud(m, buf)
			if err != nil {
				return err
			}
			{
				posid := m.NamedAttributeID(draco.GAT_POSITION)

				mesh.Verts = make([]vec3.T, node.NVert)

				var vertsSlice []float32
				vertsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&vertsSlice)))
				vertsHeader.Cap = int(node.NVert * 3)
				vertsHeader.Len = int(node.NVert * 3)
				vertsHeader.Data = uintptr(unsafe.Pointer(&mesh.Verts[0]))

				m.AttrData(m.Attr(posid), vertsSlice)
			}

			if header.Sign.Vertex.HasNormals() {
				normid := m.NamedAttributeID(draco.GAT_NORMAL)

				mesh.Normals = make([][3]int16, node.NVert)

				var normsSlice []int16
				normsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&normsSlice)))
				normsHeader.Cap = int(node.NVert * 3)
				normsHeader.Len = int(node.NVert * 3)
				normsHeader.Data = uintptr(unsafe.Pointer(&mesh.Normals[0]))

				m.AttrData(m.Attr(normid), normsSlice)
			}

			if header.Sign.Vertex.HasColors() {
				colorid := m.NamedAttributeID(draco.GAT_COLOR)

				mesh.Colors = make([][4]byte, node.NVert)

				var colorsSlice []byte
				colorsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&colorsSlice)))
				colorsHeader.Cap = int(node.NVert * 4)
				colorsHeader.Len = int(node.NVert * 4)
				colorsHeader.Data = uintptr(unsafe.Pointer(&mesh.Colors[0]))

				m.AttrData(m.Attr(colorid), colorsSlice)
			}

		} else {
			m := draco.NewMesh()
			d := draco.NewDecoder()
			err := d.DecodeMesh(m, buf)
			if err != nil {
				return err
			}
			{
				node.NFace = uint16(m.NumFaces())
				mesh.Faces = make([][3]uint16, node.NFace)

				faces := make([]uint32, node.NFace*3)
				faces = m.Faces(faces)

				for i := 0; i < int(node.NFace); i++ {
					mesh.Faces[i] = [3]uint16{uint16(faces[i*3]), uint16(faces[i*3+1]), uint16(faces[i*3+2])}
				}
			}

			{
				posid := m.NamedAttributeID(draco.GAT_POSITION)

				node.NVert = uint16(m.NumPoints())

				mesh.Verts = make([]vec3.T, node.NVert)

				var vertsSlice []float32
				vertsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&vertsSlice)))
				vertsHeader.Cap = int(node.NVert * 3)
				vertsHeader.Len = int(node.NVert * 3)
				vertsHeader.Data = uintptr(unsafe.Pointer(&mesh.Verts[0]))

				m.AttrData(m.Attr(posid), vertsSlice)
			}

			if header.Sign.Vertex.HasNormals() {
				normid := m.NamedAttributeID(draco.GAT_NORMAL)

				mesh.Normals = make([][3]int16, node.NVert)

				var normsSlice []int16
				normsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&normsSlice)))
				normsHeader.Cap = int(node.NVert * 3)
				normsHeader.Len = int(node.NVert * 3)
				normsHeader.Data = uintptr(unsafe.Pointer(&mesh.Normals[0]))

				m.AttrData(m.Attr(normid), normsSlice)
			}

			if header.Sign.Vertex.HasColors() {
				colorid := m.NamedAttributeID(draco.GAT_COLOR)

				mesh.Colors = make([][4]byte, node.NVert)

				var colorsSlice []byte
				colorsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&colorsSlice)))
				colorsHeader.Cap = int(node.NVert * 4)
				colorsHeader.Len = int(node.NVert * 4)
				colorsHeader.Data = uintptr(unsafe.Pointer(&mesh.Colors[0]))

				m.AttrData(m.Attr(colorid), colorsSlice)
			}

			if header.Sign.Vertex.HasTextures() {
				texcid := m.NamedAttributeID(draco.GAT_TEX_COORD)

				mesh.Texcoords = make([]vec2.T, node.NVert)

				var texsSlice []byte
				texsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&texsSlice)))
				texsHeader.Cap = int(node.NVert * 2)
				texsHeader.Len = int(node.NVert * 2)
				texsHeader.Data = uintptr(unsafe.Pointer(&mesh.Texcoords[0]))

				m.AttrData(m.Attr(texcid), texsSlice)
			}
		}
	}
	return nil
}

func compressNodeMesh(header Header, node *Node, mesh *NodeMesh, patches []Patch, setting *CompressSetting) NodeData {
	sig := header.Sign

	if (sig.Flags & CORTO) > 0 {
		ctx := corto.NewEncoderContext(setting.CoordQ)
		if setting.CoordBits > 0 {
			ctx.VertexBits = setting.CoordBits
		}
		ctx.NormBits = setting.NormalBits
		ctx.UvBits = float32(setting.UvBits) / 512
		ctx.ColorBits = [4]int{setting.ColorBits, setting.ColorBits, setting.ColorBits, setting.ColorBits}

		geom := &corto.Geom{}

		for p := 0; p < len(patches); p++ {
			geom.Groups = append(geom.Groups, int(patches[p].FaceOffset))
		}

		geom.Vertices = mesh.Verts[:]

		if node.NFace != 0 {
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

		buf := corto.EncodeGeom(ctx, geom)
		return buf
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

func compressTexture(header Header, img TextureImage) TextureData {
	sig := header.Sign
	if (sig.Flags & PTPNG) > 0 {
		writer := &bytes.Buffer{}
		err := png.Encode(writer, img)
		if err != nil {
			return nil
		}
		buf := writer.Bytes()
		padding := calcPadding(uint32(len(buf)), LM_PADDING)
		if padding == 0 {
			return buf
		}
		for i := 0; i < int(padding); i++ {
			buf = append(buf, byte(0))
		}
		return buf
	} else if (sig.Flags & PTJPG) > 0 {
		writer := &bytes.Buffer{}
		err := jpeg.Encode(writer, img, &jpeg.Options{Quality: LM_JPEG_QUALITY})
		if err != nil {
			return nil
		}
		buf := writer.Bytes()
		padding := calcPadding(uint32(len(buf)), LM_PADDING)
		if padding == 0 {
			return buf
		}
		for i := 0; i < int(padding); i++ {
			buf = append(buf, byte(0))
		}
		return buf
	}
	return nil
}

func decompressTexture(header Header, data TextureData) (TextureImage, error) {
	sig := header.Sign
	if (sig.Flags & PTPNG) > 0 {
		reader := bytes.NewBuffer(data)
		image, err := png.Decode(reader)
		if err != nil {
			return nil, err
		}
		return image, nil
	} else {
		reader := bytes.NewBuffer(data)
		image, err := jpeg.Decode(reader)
		if err != nil {
			return nil, err
		}
		return image, nil
	}
}
