package internal

import "sync/atomic"

type Playback struct {
	media *Media

	loop   bool
	offset int64
}

func (p *Playback) Frame() (MediaFrame, bool) {
	offset := atomic.LoadInt64(&p.offset)
	if offset == -1 {
		return MediaFrame{}, false
	}

	if offset >= int64(p.media.length) {
		if !p.loop {
			atomic.CompareAndSwapInt64(&p.offset, offset, -1)
			return MediaFrame{}, false
		}
		atomic.CompareAndSwapInt64(&p.offset, offset, 0)
		offset = 0
	}

	atomic.CompareAndSwapInt64(&p.offset, offset, offset+1)
	return MediaFrame{
		Left:  p.media.leftChannel.samples[offset],
		Right: p.media.rightChannel.samples[offset],
	}, true
}

func (p *Playback) Stop() {
	atomic.StoreInt64(&p.offset, -1)
}

func (p *Playback) IsDone() bool {
	return atomic.LoadInt64(&p.offset) == -1
}
