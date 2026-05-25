package internal

type Playback struct {
	srcNode  *PlaybackNode
	panNode  *PanNode
	gainNode *GainNode
}

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
