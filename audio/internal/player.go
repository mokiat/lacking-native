package internal

import (
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/debug/log"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

func NewPlayer() *Player {
	return &Player{}
}

type Player struct{}

func (p *Player) CreateMedia(info audio.MediaInfo) *Media {
	src, err := sdl.RWFromMem(info.Data)
	if err != nil {
		log.Error("Error staging memory: %v", err)
		return &Media{}
	}

	chunk, err := mix.LoadWAVRW(src, true)
	if err != nil {
		log.Error("Error loading chunk: %v", err)
		return &Media{}
	}

	return &Media{
		chunk: chunk,
	}
}

func (p *Player) Play(media *Media, info audio.PlayInfo) *Playback {
	chunk := media.chunk
	if chunk == nil {
		return &Playback{
			channel: -1,
		}
	}

	loopCount := 0
	if info.Loop {
		loopCount = 10000
	}

	channel, err := chunk.Play(-1, loopCount)
	if err != nil {
		log.Error("Error playing chunk: %v", err)
		return &Playback{
			channel: -1,
		}
	}

	right := uint8(255.0 * (1.0 + info.Pan) / 2.0)
	if err := mix.SetPanning(channel, 255-right, right); err != nil {
		log.Warn("Error setting pan: %v", err)
	}

	volume := int(info.Gain * 128.0)
	mix.Volume(channel, volume)

	return &Playback{
		channel: channel,
	}
}
