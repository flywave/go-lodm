package lodm

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"unsafe"

	"github.com/flywave/go3d/mat4"
)

type InstanceMesh struct {
	NodeMesh
	InstanceID  []uint32
	InstanceMat []mat4.T
}

func (m *InstanceMesh) CalcSize() int64 {
	return m.NodeMesh.CalcSize() + int64(len(m.InstanceID)*4) + int64(len(m.InstanceMat)*16*4)
}

func (m *InstanceMesh) Read(reader io.Reader, node *InstanceNode, header *Header) error {
	err := m.NodeMesh.Read(reader, &node.Node, header)

	if err != nil {
		return err
	}

	m.InstanceID = make([]uint32, node.NInstance)

	if err := binary.Read(reader, byteorder, m.InstanceID); err != nil {
		return err
	}

	m.InstanceMat = make([]mat4.T, node.NInstance)

	var matsSlice []float32
	matsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&matsSlice)))
	matsHeader.Cap = int(node.NInstance * 16)
	matsHeader.Len = int(node.NInstance * 16)
	matsHeader.Data = uintptr(unsafe.Pointer(&m.InstanceMat[0]))

	if err := binary.Read(reader, byteorder, m.InstanceMat); err != nil {
		return err
	}

	return nil
}

func (m *InstanceMesh) Write(writer io.Writer, node *InstanceNode, header *Header) error {
	err := m.NodeMesh.Write(writer, &node.Node, header)

	if err != nil {
		return err
	}

	node.NInstance = uint32(len(m.InstanceID))

	if err := binary.Write(writer, byteorder, m.InstanceID); err != nil {
		return err
	}

	var matsSlice []float32
	matsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&matsSlice)))
	matsHeader.Cap = int(node.NInstance * 16)
	matsHeader.Len = int(node.NInstance * 16)
	matsHeader.Data = uintptr(unsafe.Pointer(&m.InstanceMat[0]))

	if err := binary.Write(writer, byteorder, m.InstanceMat); err != nil {
		return err
	}
	return nil
}

func (m *InstanceMesh) getInstanceRaw() ([]byte, error) {
	si := (len(m.InstanceID) * 4) + (len(m.InstanceMat) * 16 * 4)
	ret := make([]byte, si)
	writer := bytes.NewBuffer(ret)

	ninstance := len(m.InstanceID)

	if err := binary.Write(writer, byteorder, m.InstanceID); err != nil {
		return nil, err
	}

	var matsSlice []float32
	matsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&matsSlice)))
	matsHeader.Cap = int(ninstance * 16)
	matsHeader.Len = int(ninstance * 16)
	matsHeader.Data = uintptr(unsafe.Pointer(&m.InstanceMat[0]))

	if err := binary.Write(writer, byteorder, m.InstanceMat); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

func (m *InstanceMesh) setInstanceRaw(node *InstanceNode, buf []byte) error {
	reader := bytes.NewBuffer(buf)

	m.InstanceID = make([]uint32, node.NInstance)

	if err := binary.Read(reader, byteorder, m.InstanceID); err != nil {
		return err
	}

	m.InstanceMat = make([]mat4.T, node.NInstance)

	var matsSlice []float32
	matsHeader := (*reflect.SliceHeader)((unsafe.Pointer(&matsSlice)))
	matsHeader.Cap = int(node.NInstance * 16)
	matsHeader.Len = int(node.NInstance * 16)
	matsHeader.Data = uintptr(unsafe.Pointer(&m.InstanceMat[0]))

	if err := binary.Read(reader, byteorder, m.InstanceMat); err != nil {
		return err
	}
	return nil
}

type InstanceNode struct {
	Node
	InstanceOffset uint32
	NInstance      uint32
}

func (m *InstanceNode) CalcSize() int64 {
	return int64(binary.Size(*m))
}

func (m *InstanceNode) Read(reader io.Reader) error {
	if err := binary.Read(reader, byteorder, m); err != nil {
		return err
	}
	return nil
}

func (m *InstanceNode) Write(writer io.Writer) error {
	if err := binary.Write(writer, byteorder, *m); err != nil {
		return err
	}
	return nil
}
