package lodm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
)

const (
	LM_PADDING uint32 = 256
)

var byteorder = binary.LittleEndian

type Archive struct {
	Header        Header
	Nodes         []Node
	InstanceNodes []Node
	Instances     []Instance
	Patchs        []Patch
	Textures      []Texture
	Materials     []Material
	Features      []Feature
	NodeMeshs     []NodeMesh
	InstanceMeshs []NodeMesh
	TextureImages []TextureImage
	FeatureDatas  []FeatureData

	reader  io.ReadSeekCloser
	nroots  uint32
	setting *CompressSetting
}

func NewArchive(h Header, setting *CompressSetting) *Archive {
	a := &Archive{Header: h, setting: &CompressSetting{}}
	a.optimizeCompressSetting(setting)
	return a
}

func (a *Archive) headerSize() int {
	return HeaderSize
}

func (a *Archive) indexSize() int {
	return int(a.Header.NNodes)*binary.Size(Node{}) + int(a.Header.NInstanceNodes)*binary.Size(Node{}) + int(a.Header.NInstances)*binary.Size(Instance{}) + int(a.Header.NPatches)*binary.Size(Patch{}) + int(a.Header.NTextures)*binary.Size(Texture{}) + int(a.Header.NMaterials)*binary.Size(Material{}) + int(a.Header.NFeatures)*binary.Size(Feature{})
}

func (a *Archive) initIndex() {
	a.Nodes = make([]Node, a.Header.NNodes)
	a.InstanceNodes = make([]Node, a.Header.NInstanceNodes)
	a.Instances = make([]Instance, a.Header.NInstances)
	a.Patchs = make([]Patch, a.Header.NPatches)
	a.Textures = make([]Texture, a.Header.NTextures)
	a.Materials = make([]Material, a.Header.NMaterials)
	a.Features = make([]Feature, a.Header.NFeatures)
}

func (a *Archive) countRoots() {
	a.nroots = uint32(len(a.Nodes))
	for j := 0; j < int(a.nroots); j++ {
		first_patch, last_patch := a.getNodePatchRange(uint32(j))
		for i := int(first_patch); i < int(last_patch); i++ {
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
	for i := range a.InstanceNodes {
		err = a.InstanceNodes[i].Read(a.reader)
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

func (a *Archive) getNodePatchRange(n uint32) (uint32, uint32) {
	node := &a.Nodes[n]
	nextnode := &a.Nodes[n+1]
	return node.FirstPatch, nextnode.FirstPatch - 1
}

func (a *Archive) getNodeMeshRange(n uint32) (int64, int64) {
	node := &a.Nodes[n]
	nextnode := &a.Nodes[n+1]
	return node.address(), node.address() - nextnode.address()
}

func (a *Archive) readNode(n uint32) ([]byte, error) {
	if n >= uint32(len(a.Nodes)-1) {
		return nil, errors.New("node index error")
	}
	offset, size := a.getNodeMeshRange(n)
	_, err := a.reader.Seek(offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, size)
	_, err = a.reader.Read(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *Archive) setNode(n uint32, buf []byte) error {
	sign := &a.Header.Sign
	node := &a.Nodes[n]
	nextnode := &a.Nodes[n+1]

	offset := node.address()

	d := a.NodeMeshs[n]

	compressedSize := nextnode.address() - offset

	if !sign.IsCompressed() {
		reader := bytes.NewBuffer(buf)
		err := d.Read(reader, node, &a.Header)
		if err != nil {
			return err
		}
	} else {
		err := decompressNodeMesh(buf[:compressedSize], a.Header, node, &d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) getInstanceNodePatchRange(n uint32) (uint32, uint32) {
	node := &a.InstanceNodes[n]
	nextnode := &a.InstanceNodes[n+1]
	return node.FirstPatch, nextnode.FirstPatch - 1
}

func (a *Archive) getInstanceNodeMeshRange(n uint32) (int64, int64) {
	node := &a.InstanceNodes[n]
	nextnode := &a.InstanceNodes[n+1]
	return node.address(), node.address() - nextnode.address()
}

func (a *Archive) readInstanceNode(n uint32) ([]byte, error) {
	if n >= uint32(len(a.InstanceNodes)-1) {
		return nil, errors.New("node index error")
	}
	offset, size := a.getInstanceNodeMeshRange(n)
	_, err := a.reader.Seek(offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, size)
	_, err = a.reader.Read(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *Archive) setInstanceNode(n uint32, buf []byte) error {
	sign := &a.Header.Sign
	node := &a.InstanceNodes[n]
	nextnode := &a.InstanceNodes[n+1]

	offset := node.address()

	d := a.NodeMeshs[n]

	compressedSize := nextnode.address() - offset

	if !sign.IsCompressed() {
		reader := bytes.NewBuffer(buf)
		err := d.Read(reader, node, &a.Header)
		if err != nil {
			return err
		}
	} else {
		err := decompressNodeMesh(buf[:compressedSize], a.Header, node, &d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) getPatchTextureRange(p uint32) (int64, int64) {
	t := a.Patchs[p].TexID
	tex := &a.Textures[t]
	nexttex := &a.Textures[t+1]
	return tex.address(), tex.address() - nexttex.address()
}

func (a *Archive) readPatchTexture(p uint32) ([]byte, error) {
	if p >= uint32(len(a.Patchs)-1) {
		return nil, errors.New("patch index error")
	}
	offset, size := a.getPatchTextureRange(p)
	_, err := a.reader.Seek(offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, size)
	_, err = a.reader.Read(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
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

func (a *Archive) getFeatureRange(f uint32) (int64, int64) {
	feat := &a.Features[f]
	nextfeat := &a.Features[f+1]
	return feat.address(), feat.address() - nextfeat.address()
}

func (a *Archive) readFeature(f uint32) ([]byte, error) {
	if f >= uint32(len(a.Features)-1) {
		return nil, errors.New("feature index error")
	}
	offset, size := a.getFeatureRange(f)
	_, err := a.reader.Seek(offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, size)
	_, err = a.reader.Read(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *Archive) setFeature(f uint32, buf []byte) error {
	a.FeatureDatas[f] = append(a.FeatureDatas[f], buf...)
	return nil
}

func (a *Archive) saveInstanceNodes(writer io.Writer, offset *int64) error {
	for n := 0; n < len(a.InstanceNodes); n++ {
		var nodeData NodeData
		if a.Header.Sign.IsCompressed() && a.setting != nil {
			nodeData = CompressNode(a.Header, &a.InstanceNodes[n], &a.InstanceMeshs[n], a.Patchs, a.setting)
		} else {
			buf := &bytes.Buffer{}
			a.InstanceMeshs[n].Write(buf, &a.InstanceNodes[n], &a.Header)
			nodeData = buf.Bytes()
			padding := calcPadding(uint32(len(nodeData)), LM_PADDING)
			if padding != 0 {
				for i := 0; i < int(padding); i++ {
					nodeData = append(nodeData, byte(0))
				}
			}
		}
		n, err := writer.Write(nodeData)
		if err != nil {
			return err
		}
		a.InstanceNodes[n].Offset = uint32(*offset) / LM_PADDING
		*offset += int64(n)
	}
	return nil
}

func (a *Archive) saveNodes(writer io.Writer, offset *int64) error {
	for n := 0; n < len(a.Nodes); n++ {
		var nodeData NodeData
		if a.Header.Sign.IsCompressed() && a.setting != nil {
			nodeData = CompressNode(a.Header, &a.Nodes[n], &a.NodeMeshs[n], a.Patchs, a.setting)
		} else {
			buf := &bytes.Buffer{}
			a.NodeMeshs[n].Write(buf, &a.Nodes[n], &a.Header)
			nodeData = buf.Bytes()
			padding := calcPadding(uint32(len(nodeData)), LM_PADDING)
			if padding != 0 {
				for i := 0; i < int(padding); i++ {
					nodeData = append(nodeData, byte(0))
				}
			}
		}
		n, err := writer.Write(nodeData)
		if err != nil {
			return err
		}
		a.Nodes[n].Offset = uint32(*offset) / LM_PADDING
		*offset += int64(n)
	}
	return nil
}

func (a *Archive) saveTextures(writer io.Writer, offset *int64) error {
	for i := 0; i < len(a.Textures); i++ {
		texData := compressTexture(a.Header, a.TextureImages[i])
		n, err := writer.Write(texData)
		if err != nil {
			return err
		}
		a.Textures[n].Offset = uint32(*offset) / LM_PADDING
		*offset += int64(n)
	}
	return nil
}

func (a *Archive) saveFeatures(writer io.Writer, offset *int64) error {
	for i := 0; i < len(a.Features); i++ {
		n, err := writer.Write(a.FeatureDatas[i])
		if err != nil {
			return err
		}
		a.Features[n].Offset = uint32(*offset) / LM_PADDING
		*offset += int64(n)
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
	defer writer.Close()
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
	n, err := writer.Write(tmp)
	if err != nil {
		return err
	}
	offset += int64(n)
	err = a.saveNodes(writer, &offset)
	if err != nil {
		return err
	}
	err = a.saveInstanceNodes(writer, &offset)
	if err != nil {
		return err
	}
	err = a.saveTextures(writer, &offset)
	if err != nil {
		return err
	}
	err = a.saveFeatures(writer, &offset)
	if err != nil {
		return err
	}
	writer.Seek(256, os.SEEK_SET)
	err = a.saveIndex(writer)
	if err != nil {
		return err
	}
	writer.Sync()
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
	for n := uint32(0); n < a.Header.NInstanceNodes; n++ {
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
	first_patch, last_patch := a.getInstanceNodePatchRange(n)
	t := uint32(0xffffffff)
	for p := first_patch; p < last_patch; p++ {
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
	first_patch, last_patch := a.getNodePatchRange(n)
	t := uint32(0xffffffff)
	for p := first_patch; p < last_patch; p++ {
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
	var first_patch uint32
	var last_patch uint32
	var d *NodeMesh

	if instance {
		node = &a.InstanceNodes[n]
		d = &a.InstanceMeshs[n]
		first_patch, last_patch = a.getInstanceNodePatchRange(n)
	} else {
		node = &a.Nodes[n]
		d = &a.NodeMeshs[n]
		first_patch, last_patch = a.getNodePatchRange(n)
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
		for p := first_patch; p < last_patch; p++ {
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
				for k := start; k < int(patch.FaceOffset); k++ {
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
				start = int(patch.FaceOffset)
			}
		}
	}
	return string(buffer.Bytes())
}

func (a *Archive) genNodeMtl(n uint32, instance bool) string {
	var buffer bytes.Buffer
	buffer.WriteString("# Wavefront material file\n")
	buffer.WriteString("# Converted by flywave\n")

	var first_patch uint32
	var last_patch uint32

	if instance {
		first_patch, last_patch = a.getInstanceNodePatchRange(n)
	} else {
		first_patch, last_patch = a.getNodePatchRange(n)
	}

	for p := first_patch; p < last_patch; p++ {
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
	if a.reader != nil {
		return a.reader.Close()
	}
	return nil
}

func (a *Archive) LoadAll() error {
	if a.reader != nil {
		return errors.New("file not open!")
	}
	for n := uint32(0); n < a.Header.NNodes; n++ {
		err := a.LoadNode(n)
		if err != nil {
			return err
		}
	}
	for n := uint32(0); n < a.Header.NInstanceNodes; n++ {
		err := a.LoadInstance(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) LoadNode(n uint32) error {
	if a.reader != nil {
		return errors.New("file not open!")
	}
	if !a.NodeMeshs[n].Empty() {
		return nil
	}
	nbuf, err := a.readNode(n)
	if err != nil {
		return err
	}
	err = a.setNode(n, nbuf)
	if err != nil {
		return err
	}

	first_patch, last_patch := a.getNodePatchRange(n)

	for p := first_patch; p < last_patch; p++ {
		tbuf, err := a.readPatchTexture(p)
		if err != nil {
			return err
		}
		err = a.setPatchTexture(p, tbuf)
		if err != nil {
			return err
		}
		fid := a.Patchs[p].FeatID
		fbuf, err := a.readFeature(fid)
		if err != nil {
			return err
		}
		err = a.setFeature(fid, fbuf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Archive) LoadInstance(n uint32) error {
	if a.reader != nil {
		return errors.New("file not open!")
	}

	if !a.InstanceMeshs[n].Empty() {
		return nil
	}
	nbuf, err := a.readInstanceNode(n)
	if err != nil {
		return err
	}
	err = a.setInstanceNode(n, nbuf)
	if err != nil {
		return err
	}

	first_patch, last_patch := a.getInstanceNodePatchRange(n)

	for p := first_patch; p < last_patch; p++ {
		tbuf, err := a.readPatchTexture(p)
		if err != nil {
			return err
		}
		err = a.setPatchTexture(p, tbuf)
		if err != nil {
			return err
		}
		fid := a.Patchs[p].FeatID
		fbuf, err := a.readFeature(fid)
		if err != nil {
			return err
		}
		err = a.setFeature(fid, fbuf)
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	CoordStep  float32 = 0.0
	LumaBits   int     = 6
	ChromaBits int     = 6
	AlphaBits  int     = 5
	NormBits   int     = 10
	TexStep    float32 = 0.25
)

func (a *Archive) optimizeCompressSetting(s *CompressSetting) {
	if !a.Header.Sign.IsCompressed() {
		return
	}
	coordStep := CoordStep
	if s != nil {
		if s.CoordBits > 0 {
			sphere := &a.Header.Sphere
			coordStep = sphere.Radius() / float32(math.Pow(2.0, float64(s.CoordBits)))
		}
	}

	a.setting.CoordQ = float32(math.Log2(float64(coordStep)))

	if s != nil {
		a.setting.NormalBits = s.NormalBits
		a.setting.ColorBits = s.ColorBits
		a.setting.TexStep = s.TexStep
		a.setting.UvBits = s.UvBits
	} else {
		a.setting.NormalBits = NormBits
		a.setting.ColorBits[0] = LumaBits
		a.setting.ColorBits[1] = ChromaBits
		a.setting.ColorBits[2] = ChromaBits
		a.setting.ColorBits[3] = AlphaBits
		a.setting.TexStep = TexStep
		a.setting.UvBits = int(math.Log2(float64(512 / TexStep)))
	}
}
