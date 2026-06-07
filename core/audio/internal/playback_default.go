package internal

import "github.com/mokiat/lacking/core/audio"

// DefaultPlayback is a non-spatial Processor that wraps a BasePlayback and
// registers it with its bus.
type DefaultPlayback struct {
	*BasePlayback
}

var _ audio.Playback = (*DefaultPlayback)(nil)
var _ Processor = (*DefaultPlayback)(nil)

// NewDefaultPlayback wraps basePlayback in a DefaultPlayback and registers it
// with its bus.
func NewDefaultPlayback(basePlayback *BasePlayback) *DefaultPlayback {
	result := &DefaultPlayback{
		BasePlayback: basePlayback,
	}
	basePlayback.bus.AddPlayback(result)
	return result
}

// Release removes this playback from its bus, triggering a pipeline rebuild.
func (p *DefaultPlayback) Release() {
	p.bus.RemovePlayback(p)
}
