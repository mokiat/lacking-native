package internal

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
	textures     = newMapper[*Texture]()
	samplers     = newMapper[*Sampler]()
	buffers      = newMapper[*Buffer]()
)
