package lodm

type NodeView struct {
	Node         Node
	Patchs       []Patch
	Textures     []Texture
	Materials    []Material
	Features     []Feature
	Mesh         NodeMesh
	Images       []TextureImage
	FeatureDatas []FeatureData
}

type InstanceView struct {
	NodeView
	Instances []Instance
}
