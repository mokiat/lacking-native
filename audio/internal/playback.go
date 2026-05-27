package internal

import "github.com/mokiat/lacking/audio"

type Playback struct {
	srcNode  *PlaybackNode
	panNode  *PanNode
	gainNode *GainNode
}

var _ audio.Playback = (*Playback)(nil)

func (p *Playback) Stop() {
	if p.srcNode != nil {
		p.srcNode.Stop()
	}
}

func (p *Playback) release() {
	p.srcNode.Delete()
	p.srcNode = nil

	p.panNode.Delete()
	p.panNode = nil

	p.gainNode.Delete()
	p.gainNode = nil
}
