package internal

// TODO: Experiment with using a slice as a mapping mechanism
// instead of a map.
// It is possible to use a freeIDs slice to keep track of
// freed up slots so that the array need not be iterated.

func newMapper[T any]() *Mapper[T] {
	return &Mapper[T]{
		mapping: make(map[uint32]T),
	}
}

type Mapper[T any] struct {
	mapping map[uint32]T
}

func (m *Mapper[T]) Track(id uint32, v T) {
	m.mapping[id] = v
}

func (m *Mapper[T]) Release(id uint32) {
	delete(m.mapping, id)
}

func (m *Mapper[T]) Get(id uint32) T {
	return m.mapping[id]
}

var (
	framebuffers = newMapper[*Framebuffer]()
	programs     = newMapper[*Program]()
	textures     = newMapper[*Texture]()
	buffers      = newMapper[*Buffer]()
	vertexArrays = newMapper[*VertexArray]()
)
