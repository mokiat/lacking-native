package internal

import "sync/atomic"

type Playback struct {
	media *Media

	loop   bool
	offset int64
}

func (p *Playback) Frame() (mediaSample, bool) {
	offset := atomic.LoadInt64(&p.offset)
	if offset == -1 {
		return mediaSample{}, false
	}
	if offset >= int64(len(p.media.frames)) {
		if p.loop {
			atomic.CompareAndSwapInt64(&p.offset, offset, 0)
			offset = 0
		} else {
			atomic.CompareAndSwapInt64(&p.offset, offset, -1)
			return mediaSample{}, false
		}
	}
	atomic.CompareAndSwapInt64(&p.offset, offset, offset+1)
	return p.media.frames[offset], true
}

func (p *Playback) Stop() {
	atomic.StoreInt64(&p.offset, -1)
}

func (p *Playback) IsDone() bool {
	return atomic.LoadInt64(&p.offset) == -1
}
