package lodm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"reflect"
	"unsafe"

	"github.com/flywave/go-corto"
	"github.com/flywave/go-draco"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

const (
	LM_PADDING   = 256
	JPEG_QUALITY = 70
)

var byteorder = binary.LittleEndian

type Archive struct {
	Header        Header
	Nodes         []Node
	Instances     []InstanceNode
	Patchs        []Patch
	Textures      []Texture
	Materials     []Material
	Features      []Feature
	NodeMeshs     []NodeMesh
	InstanceMeshs []InstanceMesh
	TextureImages []TextureImage
	FeatureDatas  []FeatureData

	reader io.ReadSeekCloser
	path   string
	size   int
	nroots uint32
}

func (a *Archive) indexSize() int {
	return int(a.Header.NNodes)*binary.Size(Node{}) + int(a.Header.NInstances)*binary.Size(InstanceNode{}) + int(a.Header.NPatches)*binary.Size(Patch{}) + int(a.Header.NTextures)*binary.Size(Texture{}) + int(a.Header.NMaterials)*binary.Size(Material{}) + int(a.Header.NFeatures)*binary.Size(Feature{})
}

func (a *Archive) initIndex() {
	a.Nodes = make([]Node, a.Header.NNodes)
	a.Instances = make([]InstanceNode, a.Header.NInstances)
	a.Patchs = make([]Patch, a.Header.NPatches)
	a.Textures = make([]Texture, a.Header.NTextures)
	a.Materials = make([]Material, a.Header.NMaterials)
	a.Features = make([]Feature, a.Header.NFeatures)
}

func (a *Archive) countRoots() {
	a.nroots = uint32(len(a.Nodes))
	for j := 0; j < int(a.nroots); j++ {
		for i := int(a.Nodes[j].FirstPatch); i < int(a.Nodes[j+1].FirstPatch); i++ {
			if a.Patchs[i].Node < uint32(a.nroots) {
				a.nroots = uint32(a.Patchs[i].Node)
			}
		}
	}
}

func (a *Archive) loadHeader() error {
	err := a.Header.Read(a.reader)
	return err
}

func (a *Archive) loadIndex() error {
	a.initIndex()
	var err error
	for i := range a.Nodes {
		err = a.Nodes[i].Read(a.reader)
		if err != nil {
			return err
		}
	}
	for i := range a.Instances {
		err = a.Instances[i].Read(a.reader)
		if err != nil {
			return err
		}
	}
	for i := range a.Patchs {
		err = a.Patchs[i].Read(a.reader)
		if err != nil {
			return err
		}
	}
	for i := range a.Textures {
		err = a.Textures[i].Read(a.reader)
		if err != nil {
			return err
		}
	}
	for i := range a.Materials {
		err = a.Materials[i].Read(a.reader)
		if err != nil {
			return err
		}
	}
	for i := range a.Features {
		err = a.Features[i].Read(a.reader)
		if err != nil {
			return err
		}
	}
	a.countRoots()
	return nil
}

func (a *Archive) loadNode(n uint32) []byte {
	return nil
}

func (a *Archive) setNode(n uint32, buf []byte) error {
	sign := &a.Header.Sign
	node := &a.Nodes[n]
	nextnode := &a.Nodes[n+1]

	offset := node.Offset

	d := a.NodeMeshs[n]

	compressedSize := nextnode.Offset - offset

	if !sign.IsCompressed() {
		reader := bytes.NewBuffer(buf)
		err := d.Read(reader, node, &a.Header)
		if err != nil {
			return err
		}
	} else {
		decompressNodeMesh(buf[:compressedSize], a.Header, node, &d)
	}
	return nil
}

func (a *Archive) loadInstance(p uint32) []byte {
	return nil
}

func (a *Archive) setInstanceNode(n uint32, buf []byte) error {
	sign := &a.Header.Sign
	node := &a.Instances[n]
	nextnode := &a.Instances[n+1]

	offset := node.Offset

	d := a.InstanceMeshs[n]

	compressedSize := nextnode.Offset - offset

	if !sign.IsCompressed() {
		reader := bytes.NewBuffer(buf)
		err := d.Read(reader, node, &a.Header)
		if err != nil {
			return err
		}
	} else {
		decompressNodeMesh(buf[:compressedSize], a.Header, &node.Node, &d.NodeMesh)
		d.setInstanceRaw(node, buf[node.InstanceOffset:])
	}
	return nil
}

func (a *Archive) loadTexture(p uint32) []byte {
	return nil
}

func (a *Archive) setPatchTexture(p uint32, buf []byte) error {
	t := a.Patchs[p].TexID
	if t == 0xffffffff {
		return nil
	}
	var err error
	a.TextureImages[t], err = decompressTexture(a.Header, buf)
	if err != nil {
		return err
	}
	return nil
}

func (a *Archive) loadFeature(f uint32) []byte {
	return nil
}

func (a *Archive) setFeature(f uint32, buf []byte) error {
	a.FeatureDatas[f] = append(a.FeatureDatas[f], buf...)
	return nil
}

func (a *Archive) saveNodes(writer io.Writer) error {
	for i := 0; i < len(a.NodeMeshs); i++ {

	}
	return nil
}

func (a *Archive) saveInstances(writer io.Writer) error {
	for i := 0; i < len(a.InstanceMeshs); i++ {

	}
	return nil
}

func (a *Archive) saveTextures(writer io.Writer) error {
	for i := 0; i < len(a.TextureImages); i++ {

	}
	return nil
}

func (a *Archive) saveFeatures(writer io.Writer) error {
	for i := 0; i < len(a.FeatureDatas); i++ {

	}
	return nil
}

func (a *Archive) saveHeader(writer io.Writer) error {
	return a.Header.Write(writer)
}

func (a *Archive) saveIndex(writer io.Writer) error {
	var err error
	for i := range a.Nodes {
		err = a.Nodes[i].Write(writer)
		if err != nil {
			return err
		}
	}
	for i := range a.Instances {
		err = a.Instances[i].Write(writer)
		if err != nil {
			return err
		}
	}
	for i := range a.Patchs {
		err = a.Patchs[i].Write(writer)
		if err != nil {
			return err
		}
	}
	for i := range a.Textures {
		err = a.Textures[i].Write(writer)
		if err != nil {
			return err
		}
	}
	for i := range a.Materials {
		err = a.Materials[i].Write(writer)
		if err != nil {
			return err
		}
	}
	for i := range a.Features {
		err = a.Features[i].Write(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) BoundingSpere() Sphere {
	return a.Header.Sphere
}

func (a *Archive) Open(path string) error {
	var err error
	a.reader, err = os.Open(path)
	if err != nil {
		return err
	}
	err = a.loadHeader()
	if err != nil {
		return err
	}
	err = a.loadIndex()
	if err != nil {
		return err
	}
	return nil
}

func (a *Archive) Save(path string) error {
	writer, err := os.Open(path)
	if err != nil {
		return err
	}
	err = a.saveHeader(writer)
	if err != nil {
		return err
	}
	err = a.saveIndex(writer)
	if err != nil {
		return err
	}
	offset, _ := writer.Seek(0, os.SEEK_CUR)
	padding := calcPadding(uint32(offset), LM_PADDING)
	tmp := make([]byte, padding)
	_, err = writer.Write(tmp)
	if err != nil {
		return err
	}
	err = a.saveNodes(writer)
	if err != nil {
		return err
	}
	err = a.saveInstances(writer)
	if err != nil {
		return err
	}
	err = a.saveTextures(writer)
	if err != nil {
		return err
	}
	err = a.saveFeatures(writer)
	if err != nil {
		return err
	}
	return nil
}

func (a *Archive) Extract(path_ string) error {
	for n := uint32(0); n < a.Header.NNodes; n++ {
		obj := a.genNodeObj(n, false)
		objname := fmt.Sprintf("node_%v.obj", n)

		f, err := os.Open(path.Join(path_, objname))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(obj))
		if err != nil {
			return err
		}

		mtl := a.genNodeMtl(n, false)
		mtlname := fmt.Sprintf("node_%v.mtl", n)

		f, err = os.Open(path.Join(path_, mtlname))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(mtl))
		if err != nil {
			return err
		}

		err = a.extractNodeTexture(path_, n)
		if err != nil {
			return err
		}
	}
	for n := uint32(0); n < a.Header.NInstances; n++ {
		obj := a.genNodeObj(n, true)
		objname := fmt.Sprintf("instance_%v.obj", n)

		f, err := os.Open(path.Join(path_, objname))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(obj))
		if err != nil {
			return err
		}

		mtl := a.genNodeMtl(n, true)
		mtlname := fmt.Sprintf("instance_%v.mtl", n)

		f, err = os.Open(path.Join(path_, mtlname))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(mtl))
		if err != nil {
			return err
		}

		err = a.extractInstanceNodeTexture(path_, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) extractInstanceNodeTexture(path_ string, n uint32) error {
	node := &a.Instances[n]
	last_patch := a.Instances[n+1].FirstPatch - 1
	t := uint32(0xffffffff)
	for p := node.FirstPatch; p < last_patch; p++ {
		t_ := a.Patchs[p].TexID
		if t_ == 0xffffffff || t_ == t {
			continue
		}
		t = t_
		var filename string
		if (a.Header.Sign.Flags & PTJPG) > 0 {
			filename = fmt.Sprintf("instance_%v_%v_tex.jpg", n, t)
		} else if (a.Header.Sign.Flags & PTPNG) > 0 {
			filename = fmt.Sprintf("instance_%v_%v_tex.png", n, t)
		}
		texbuf := a.TextureImages[t]
		texdata := compressTexture(a.Header, texbuf)
		f, err := os.Open(path.Join(path_, filename))
		if err != nil {
			return err
		}
		_, err = f.Write(texdata)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) extractNodeTexture(path_ string, n uint32) error {
	node := &a.Nodes[n]
	last_patch := a.Nodes[n+1].FirstPatch - 1
	t := uint32(0xffffffff)
	for p := node.FirstPatch; p < last_patch; p++ {
		t_ := a.Patchs[p].TexID
		if t_ == 0xffffffff || t_ == t {
			continue
		}
		t = t_
		var filename string
		if (a.Header.Sign.Flags & PTJPG) > 0 {
			filename = fmt.Sprintf("node_%v_%v_tex.jpg", n, t)
		} else if (a.Header.Sign.Flags & PTPNG) > 0 {
			filename = fmt.Sprintf("node_%v_%v_tex.png", n, t)
		}
		texbuf := a.TextureImages[t]
		texdata := compressTexture(a.Header, texbuf)
		f, err := os.Open(path.Join(path_, filename))
		if err != nil {
			return err
		}
		_, err = f.Write(texdata)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) genNodeObj(n uint32, instance bool) string {
	var buffer bytes.Buffer
	buffer.WriteString("# Wavefront OBJ file\n")
	buffer.WriteString("# Converted by flywave\n")
	if instance {
		buffer.WriteString(fmt.Sprintf("mtllib instance_%v.mtl\n", n))
	} else {
		buffer.WriteString(fmt.Sprintf("mtllib node_%v.mtl\n", n))
	}
	sign := &a.Header.Sign
	var node *Node
	var last_patch uint32
	var d *NodeMesh
	if instance {
		node = &a.Instances[n].Node
		last_patch = a.Instances[n+1].FirstPatch - 1
		d = &a.InstanceMeshs[n].NodeMesh
	} else {
		node = &a.Nodes[n]
		last_patch = a.Nodes[n+1].FirstPatch - 1
		d = &a.NodeMeshs[n]
	}

	buffer.WriteString(fmt.Sprintf("# object %v\n", n))
	for i := 0; i < int(node.NVert); i++ {
		buffer.WriteString(fmt.Sprintf("v %f %f %f\n", d.Verts[i][0], d.Verts[i][1], d.Verts[i][2]))
	}
	buffer.WriteString(fmt.Sprintf("#  %v vertices\n", node.NVert))

	if sign.Vertex.HasNormals() {
		for i := 0; i < int(node.NVert); i++ {
			buffer.WriteString(fmt.Sprintf("vn %d %d %d\n", d.Normals[i][0], d.Normals[i][1], d.Normals[i][2]))
		}
		buffer.WriteString(fmt.Sprintf("#  %v normals\n", node.NVert))
	}

	if sign.Vertex.HasTextures() {
		for i := 0; i < int(node.NVert); i++ {
			buffer.WriteString(fmt.Sprintf("vt %f %f \n", d.Texcoords[i][0], d.Texcoords[i][1]))
		}
		buffer.WriteString(fmt.Sprintf("#  %v texture vertices\n", node.NVert))
	}

	if node.NFace > 0 {
		start := 0
		for p := node.FirstPatch; p < last_patch; p++ {
			buffer.WriteString(fmt.Sprintf("g patch-%v\n", p))

			patch := &a.Patchs[p]
			t := a.Patchs[p].TexID
			m := a.Patchs[p].MtlID

			if t != 0xffffffff || m != 0xffffffff {
				if instance {
					buffer.WriteString(fmt.Sprintf("usemtl instance_%v_%v_%v\n", n, t, m))
				} else {
					buffer.WriteString(fmt.Sprintf("usemtl node_%v_%v_%v\n", n, t, m))
				}
				for k := start; k < int(patch.Offset); k++ {
					fline := "f "
					fline += fmt.Sprintf("%d", (d.Faces[k][0] + 1))
					if sign.Vertex.HasTextures() {
						fline += fmt.Sprintf("/%d", (d.Faces[k][0] + 1))
					} else {
						if sign.Vertex.HasNormals() {
							fline += "/"
						}
					}

					if sign.Vertex.HasNormals() {
						fline += fmt.Sprintf("/%d", (d.Faces[k][0] + 1))
					}
					fline += " "

					fline += fmt.Sprintf("%d", (d.Faces[k][1] + 1))
					if sign.Vertex.HasTextures() {
						fline += fmt.Sprintf("/%d", (d.Faces[k][1] + 1))
					} else {
						if sign.Vertex.HasNormals() {
							fline += "/"
						}
					}

					if sign.Vertex.HasNormals() {
						fline += fmt.Sprintf("/%d", (d.Faces[k][1] + 1))
					}
					fline += " "

					fline += fmt.Sprintf("%d", (d.Faces[k][2] + 1))
					if sign.Vertex.HasTextures() {
						fline += fmt.Sprintf("/%d", (d.Faces[k][2] + 1))
					} else {
						if sign.Vertex.HasNormals() {
							fline += "/"
						}
					}

					if sign.Vertex.HasNormals() {
						fline += fmt.Sprintf("/%d", (d.Faces[k][2] + 1))
					}
					fline += "\n"

					buffer.WriteString(fline)
				}
				start = int(patch.Offset)
			}
		}
	}
	return string(buffer.Bytes())
}

func (a *Archive) genNodeMtl(n uint32, instance bool) string {
	var buffer bytes.Buffer
	buffer.WriteString("# Wavefront material file\n")
	buffer.WriteString("# Converted by flywave\n")

	var node *Node
	var last_patch uint32
	if instance {
		node = &a.Instances[n].Node
		last_patch = a.Instances[n+1].FirstPatch - 1
	} else {
		node = &a.Nodes[n]
		last_patch = a.Nodes[n+1].FirstPatch - 1
	}

	for p := node.FirstPatch; p < last_patch; p++ {
		t := a.Patchs[p].TexID
		m := a.Patchs[p].MtlID

		if t == 0xffffffff || m == 0xffffffff {
			continue
		}
		if instance {
			buffer.WriteString(fmt.Sprintf("newmtl instance_%v_%v_%v\n", n, t, m))
		} else {
			buffer.WriteString(fmt.Sprintf("newmtl node_%v_%v_%v\n", n, t, m))
		}
		if m != 0xffffffff {
			mat := a.Materials[m]

			buffer.WriteString(fmt.Sprintf("Kd %v %v %v \n", mat.Color[0]/255, mat.Color[1]/255, mat.Color[2]/255))
			buffer.WriteString(fmt.Sprintf("d %v \n", mat.Opacity))

			if mat.Type == MTL_LAMBERT || mat.Type == MTL_PHONG {
				buffer.WriteString(fmt.Sprintf("Ka %v %v %v \n", mat.Ambient[0]/255, mat.Ambient[1]/255, mat.Ambient[2]/255))
			}

			if mat.Type == MTL_PHONG {
				buffer.WriteString(fmt.Sprintf("Ks %v %v %v \n", mat.Specular[0]/255, mat.Specular[1]/255, mat.Specular[2]/255))
				buffer.WriteString(fmt.Sprintf("Ns %v \n", mat.Shininess))
			} else {
				buffer.WriteString("Ks 0.000000 0.000000 0.000000\n")
				buffer.WriteString("Ns 8.000000\n")
			}
			buffer.WriteString("illum 2\n")
		} else {
			buffer.WriteString("Ka 0.000000 0.000000 0.000000\n")
			buffer.WriteString("Kd 1.000000 1.000000 1.000000\n")
			buffer.WriteString("Ks 0.000000 0.000000 0.000000\n")
			buffer.WriteString("illum 2\n")
			buffer.WriteString("Ns 8.000000\n")
		}
		if t != 0xffffffff {
			if instance {
				buffer.WriteString(fmt.Sprintf("map_Kd instance_%v_%v_tex.jpg \n", n, t))
			} else {
				buffer.WriteString(fmt.Sprintf("map_Kd node_%v_%v_tex.jpg \n", n, t))
			}
		}
	}
	return string(buffer.Bytes())
}

func (a *Archive) Close() error {
	return a.reader.Close()
}

func (a *Archive) LoadAll() error {
	for n := uint32(0); n < a.Header.NNodes; n++ {
		err := a.LoadNode(n)
		if err != nil {
			return err
		}
	}
	for n := uint32(0); n < a.Header.NInstances; n++ {
		err := a.LoadInstance(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) LoadNode(n uint32) error {
	if !a.NodeMeshs[n].Empty() {
		return nil
	}
	nbuf := a.loadNode(n)
	err := a.setNode(n, nbuf)
	if err != nil {
		return err
	}

	node := &a.Nodes[n]
	last_patch := a.Nodes[n+1].FirstPatch - 1

	for p := node.FirstPatch; p < last_patch; p++ {
		tbuf := a.loadTexture(p)
		a.setPatchTexture(p, tbuf)

		fid := a.Patchs[p].FeatID
		fbuf := a.loadFeature(fid)
		a.setFeature(fid, fbuf)
	}

	return nil
}

func (a *Archive) LoadInstance(n uint32) error {
	if !a.InstanceMeshs[n].Empty() {
		return nil
	}
	nbuf := a.loadInstance(n)
	err := a.setInstanceNode(n, nbuf)
	if err != nil {
		return err
	}

	node := &a.Instances[n]
	last_patch := a.Instances[n+1].FirstPatch - 1

	for p := node.FirstPatch; p < last_patch; p++ {
		tbuf := a.loadTexture(p)
		a.setPatchTexture(p, tbuf)

		fid := a.Patchs[p].FeatID
		fbuf := a.loadFeature(fid)
		a.setFeature(fid, fbuf)
	}

	return nil
}

type CompressSetting struct {
	CoordQ     int
	CoordBits  int
	NormalBits int
	UvBits     int
	ColorBits  int
}

func CompressNode(header Header, node *Node, mesh *NodeMesh, patches []Patch, setting CompressSetting) NodeData {
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

func CompressInstanceNode(header Header, node *InstanceNode, mesh *InstanceMesh, patches []Patch, setting CompressSetting) NodeData {
	buf := compressNodeMesh(header, &node.Node, &mesh.NodeMesh, patches, setting)
	node.InstanceOffset = uint32(len(buf))
	node.NInstance = uint32(len(mesh.InstanceID))
	ibuf := mesh.getInstanceRaw()
	buf = append(buf, ibuf...)
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
			posid := m.NamedAttributeID(draco.GAT_POSITION)

			mesh.Verts = make([]vec3.T, node.NVert)

			var vertsSlice []float32
			vertsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&vertsSlice)))
			vertsHeader.Cap = int(node.NVert * 3)
			vertsHeader.Len = int(node.NVert * 3)
			vertsHeader.Data = uintptr(unsafe.Pointer(&mesh.Verts[0]))

			m.AttrData(m.Attr(posid), vertsSlice)

		} else {
			m := draco.NewMesh()
			d := draco.NewDecoder()
			err := d.DecodeMesh(m, buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func compressNodeMesh(header Header, node *Node, mesh *NodeMesh, patches []Patch, setting CompressSetting) NodeData {
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
		buf := make([]byte, img.Bounds().Dx()*img.Bounds().Dy()*4)
		writer := bytes.NewBuffer(buf)
		err := png.Encode(writer, img)
		if err != nil {
			return nil
		}
		buf = writer.Bytes()
		padding := calcPadding(uint32(len(buf)), LM_PADDING)
		if padding == 0 {
			return buf
		}
		for i := 0; i < int(padding); i++ {
			buf = append(buf, byte(0))
		}
		return buf
	} else if (sig.Flags & PTJPG) > 0 {
		buf := make([]byte, img.Bounds().Dx()*img.Bounds().Dy()*4)
		writer := bytes.NewBuffer(buf)
		err := jpeg.Encode(writer, img, &jpeg.Options{Quality: JPEG_QUALITY})
		if err != nil {
			return nil
		}
		buf = writer.Bytes()
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
