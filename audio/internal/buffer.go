package internal

import "slices"

func newBuffer(initialCapacity uint32) *Buffer {
	return &Buffer{
		frames:    make([]Frame, initialCapacity),
		offset:    0,
		partition: 32,
	}
}

type Buffer struct {
	frames    []Frame
	offset    uint32
	partition uint32
}

func (b *Buffer) Reset(partition uint32) {
	b.offset = 0
	b.partition = partition
}

func (b *Buffer) Allocate() []Frame {
	// Ensure sufficient space.
	if gap := int(b.offset+b.partition) - len(b.frames); gap > 0 {
		b.frames = slices.Grow(b.frames, gap)
		b.frames = b.frames[:cap(b.frames)]
	}
	// Issue the partition.
	result := b.frames[b.offset : b.offset+b.partition]
	clear(result)
	b.offset += b.partition
	return result
}
