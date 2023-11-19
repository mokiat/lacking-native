package internal

import "github.com/veandco/go-sdl2/mix"

type Playback struct {
	channel int
}

func (p *Playback) Stop() {
	if p.channel >= 0 {
		mix.HaltChannel(p.channel)
		p.channel = -1
	}
}
