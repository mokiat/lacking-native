package internal

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/core/audio"
)

type MasterBus struct {
	buses *ds.List[Bus]

	gainFilter        *GainFilter
	compressionFilter *CompressionFilter
}

var _ audio.MasterBus = (*MasterBus)(nil)
var _ Processor = (*MasterBus)(nil)

func NewMasterBus() *MasterBus {
	return &MasterBus{
		buses: ds.EmptyList[Bus](),

		gainFilter:        NewGainFilter(),
		compressionFilter: NewCompressionFilter(),
	}
}

func (b *MasterBus) Gain() float32 {
	return b.gainFilter.Gain()
}

func (b *MasterBus) SetGain(gain float32) {
	b.gainFilter.SetGain(gain)
}

func (b *MasterBus) Compression() audio.Compression {
	return b.compressionFilter
}

func (b *MasterBus) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	outputFrames := b.gainFilter.Process(ctx, inputFrames)
	outputFrames = b.compressionFilter.Process(ctx, outputFrames)
	return outputFrames
}
