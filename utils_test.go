package lodm

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

func TestCompressDecompressTexture(t *testing.T) {
	image := image.NewNRGBA(image.Rect(0, 0, 256, 256))

	for x := 0; x < 256; x++ {
		for y := 0; y < 256; y++ {
			image.Set(x, y, &color.NRGBA{255, 128, 255, 255})
		}
	}

	s := Signature{}
	s.SetFlag(PTJPG)

	h := NewHeader(s)

	data := compressTexture(*h, image)

	if len(data) == 0 {
		t.FailNow()
	}

	tex, err := decompressTexture(*h, data)

	if err != nil || !tex.Bounds().Eq(image.Bounds()) {
		t.FailNow()
	}

	s = Signature{}
	s.SetFlag(PTPNG)

	h = NewHeader(s)

	data = compressTexture(*h, image)

	if len(data) == 0 {
		t.FailNow()
	}

	tex, err = decompressTexture(*h, data)

	if err != nil || !tex.Bounds().Eq(image.Bounds()) {
		t.FailNow()
	}

}

var (
	testPC = NodeMesh{
		Verts: []vec3.T{
			{0., 0., 0.},
			{1., 0., 0.},
			{0., 1., 0.},
			{0., 1., 0.},
			{1., 0., 0.},
			{1., 1., 0.},
			{0., 1., 1.},
			{1., 0., 1.},
			{0., 0., 1.},
			{1., 1., 1.},
			{1., 0., 1.},
			{0., 1., 1.},
			{0., 1., 0.},
			{1., 1., 0.},
			{0., 1., 1.},
			{0., 1., 1.},
			{1., 1., 0.},
			{1., 1., 1.},
			{0., 0., 1.},
			{1., 0., 0.},
			{0., 0., 0.},
			{1., 0., 1.},
			{1., 0., 0.},
			{0., 0., 1.},
			{1., 0., 0.},
			{1., 0., 1.},
			{1., 1., 0.},
			{1., 1., 0.},
			{1., 0., 1.},
			{1., 1., 1.},
			{0., 1., 0.},
			{0., 0., 1.},
			{0., 0., 0.},
			{0., 1., 1.},
			{0., 0., 1.},
			{0., 1., 0.},
		},
		Normals: [][3]int16{
			{0., 0., 0.},
			{1., 0., 0.},
			{0., 1., 0.},
			{0., 1., 0.},
			{1., 0., 0.},
			{1., 1., 0.},
			{0., 1., 1.},
			{1., 0., 1.},
			{0., 0., 1.},
			{1., 1., 1.},
			{1., 0., 1.},
			{0., 1., 1.},
			{0., 1., 0.},
			{1., 1., 0.},
			{0., 1., 1.},
			{0., 1., 1.},
			{1., 1., 0.},
			{1., 1., 1.},
			{0., 0., 1.},
			{1., 0., 0.},
			{0., 0., 0.},
			{1., 0., 1.},
			{1., 0., 0.},
			{0., 0., 1.},
			{1., 0., 0.},
			{1., 0., 1.},
			{1., 1., 0.},
			{1., 1., 0.},
			{1., 0., 1.},
			{1., 1., 1.},
			{0., 1., 0.},
			{0., 0., 1.},
			{0., 0., 0.},
			{0., 1., 1.},
			{0., 0., 1.},
			{0., 1., 0.},
		},
		Colors: [][4]byte{
			{128, 128, 128, 255},
			{255, 128, 128, 255},
			{128, 255, 128, 255},
			{128, 255, 128, 255},
			{255, 128, 128, 255},
			{255, 255, 128, 255},
			{128, 255, 255, 255},
			{255, 128, 255, 255},
			{128, 128, 255, 255},
			{255, 255, 255, 255},
			{255, 128, 255, 255},
			{128, 255, 255, 255},
			{128, 255, 128, 255},
			{255, 255, 128, 255},
			{128, 255, 255, 255},
			{128, 255, 255, 255},
			{255, 255, 128, 255},
			{255, 255, 255, 255},
			{128, 128, 255, 255},
			{255, 128, 128, 255},
			{128, 128, 128, 255},
			{255, 128, 255, 255},
			{255, 128, 128, 255},
			{128, 128, 255, 255},
			{255, 128, 128, 255},
			{255, 128, 255, 255},
			{255, 255, 128, 255},
			{255, 255, 128, 255},
			{255, 128, 255, 255},
			{255, 255, 255, 255},
			{128, 255, 128, 255},
			{128, 128, 255, 255},
			{128, 128, 128, 255},
			{128, 255, 255, 255},
			{128, 128, 255, 255},
			{128, 255, 128, 255},
		},
	}
	testMesh = NodeMesh{
		Verts: []vec3.T{
			{0., 0., 0.},
			{1., 0., 0.},
			{0., 1., 0.},
			{1., 1., 0.},
			{0., 1., 1.},
			{1., 0., 1.},
			{0., 0., 1.},
			{1., 1., 1.},
			{1., 1., 0.},
		},
		Faces: [][3]uint16{
			{0, 1, 2},
			{2, 1, 3},
			{4, 5, 6},
			{7, 5, 4},
			{2, 3, 4},
			{4, 8, 7},
			{6, 1, 0},
			{5, 1, 6},
			{1, 5, 8},
			{8, 5, 7},
			{2, 6, 0},
			{4, 6, 2},
		},
	}
	testMesh2 = NodeMesh{
		Verts: []vec3.T{
			{0., 0., 0.},
			{1., 0., 0.},
			{0., 1., 0.},
			{0., 1., 0.},
			{1., 0., 0.},
			{1., 1., 0.},
			{0., 1., 1.},
			{1., 0., 1.},
			{0., 0., 1.},
			{1., 1., 1.},
			{1., 0., 1.},
			{0., 1., 1.},
			{0., 1., 0.},
			{1., 1., 0.},
			{0., 1., 1.},
			{0., 1., 1.},
			{1., 1., 0.},
			{1., 1., 1.},
			{0., 0., 1.},
			{1., 0., 0.},
			{0., 0., 0.},
			{1., 0., 1.},
			{1., 0., 0.},
			{0., 0., 1.},
			{1., 0., 0.},
			{1., 0., 1.},
			{1., 1., 0.},
			{1., 1., 0.},
			{1., 0., 1.},
			{1., 1., 1.},
			{0., 1., 0.},
			{0., 0., 1.},
			{0., 0., 0.},
			{0., 1., 1.},
			{0., 0., 1.},
			{0., 1., 0.}},
		Faces: [][3]uint16{
			{0, 1, 2},
			{3, 4, 5},
			{6, 7, 8},
			{9, 10, 11},
			{12, 13, 14},
			{15, 16, 17},
			{18, 19, 20},
			{21, 22, 23},
			{24, 25, 26},
			{27, 28, 29},
			{30, 31, 32},
			{33, 34, 35},
		},
		Texcoords: []vec2.T{
			{0, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
			{0.5, 0.5},
		},
	}
)

func TestDracoMesh(t *testing.T) {
	sign := &Signature{}

	sign.Vertex.SetComponent(VERTEX_COORD, Attribute{Type: ATTR_FLOAT, Number: 3})
	sign.Vertex.SetComponent(VERTEX_TEX, Attribute{Type: ATTR_FLOAT, Number: 2})

	sign.Face.SetComponent(FACE_INDEX, Attribute{Type: ATTR_UNSIGNED_SHORT, Number: 3})

	sign.SetFlag(DRACO)

	h := NewHeader(*sign)

	node := &Node{NVert: uint16(len(testMesh2.Verts)), NFace: uint16(len(testMesh2.Faces))}

	data := CompressNode(*h, node, &testMesh2, nil, &DEFAULE_COMPRESS_SETTING)

	if len(data) == 0 {
		t.FailNow()
	}

	var mesh NodeMesh
	err := decompressNodeMesh(data, *h, node, &mesh)

	if err != nil {
		t.FailNow()
	}

	if len(mesh.Faces) != len(testMesh2.Faces) {
		t.FailNow()
	}
}

func TestCortoMesh(t *testing.T) {
	sign := &Signature{}

	sign.Vertex.SetComponent(VERTEX_COORD, Attribute{Type: ATTR_FLOAT, Number: 3})
	sign.Vertex.SetComponent(VERTEX_TEX, Attribute{Type: ATTR_FLOAT, Number: 2})

	sign.Face.SetComponent(FACE_INDEX, Attribute{Type: ATTR_UNSIGNED_SHORT, Number: 3})

	sign.SetFlag(CORTO)

	h := NewHeader(*sign)

	node := &Node{NVert: uint16(len(testMesh2.Verts)), NFace: uint16(len(testMesh2.Faces))}

	data := CompressNode(*h, node, &testMesh2, nil, &DEFAULE_COMPRESS_SETTING)

	if len(data) == 0 {
		t.FailNow()
	}

	var mesh NodeMesh
	err := decompressNodeMesh(data, *h, node, &mesh)

	if err != nil {
		t.FailNow()
	}

	if len(mesh.Faces) != len(testMesh2.Faces) {
		t.FailNow()
	}
}

func TestSize(t *testing.T) {
	si := binary.Size(Patch{})
	fmt.Printf("Patch-Size: %v", si)
	si = binary.Size(Node{})
	fmt.Printf("Node-Size: %v", si)
	si = binary.Size(Instance{})
	fmt.Printf("Instance-Size: %v", si)
	si = binary.Size(Material{})
	fmt.Printf("Material-Size: %v", si)
	si = binary.Size(Texture{})
	fmt.Printf("Texture-Size: %v", si)
	si = binary.Size(Feature{})
	fmt.Printf("Feature-Size: %v", si)
}
