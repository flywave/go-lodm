package lodm

import "github.com/flywave/go3d/vec3"

func calcPadding(offset, paddingUnit uint32) uint32 {
	padding := offset % paddingUnit
	if padding != 0 {
		padding = paddingUnit - padding
	}
	return padding
}

func paddingBytes(bytes []byte, srcLen int, paddingUnit uint32, paddingCode byte) {
	padding := calcPadding(uint32(srcLen), paddingUnit)

	for i := 0; i < int(padding); i++ {
		bytes[(srcLen)+i] = paddingCode
	}
}

func createPaddingBytes(bytes []byte, offset, paddingUnit uint32, paddingCode byte) []byte {
	padding := calcPadding(offset, paddingUnit)
	if padding == 0 {
		return bytes
	}
	for i := 0; i < int(padding); i++ {
		bytes = append(bytes, paddingCode)
	}
	return bytes
}

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
