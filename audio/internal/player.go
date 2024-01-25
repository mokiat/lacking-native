package internal

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/hajimehoshi/go-mp3"
	"github.com/mokiat/gblob"
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/debug/log"
)

func NewPlayer() (*Player, error) {
	player := &Player{
		playbacks: make(map[*Playback]struct{}),
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating malgo context: %w", err)
	}
	player.ctx = ctx

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = 44100
	deviceConfig.Alsa.NoMMap = 1

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: player.onSamples,
		Stop: player.onStop,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return nil, fmt.Errorf("error creating malgo device: %w", err)
	}
	player.device = device

	if err := device.Start(); err != nil {
		return nil, fmt.Errorf("error starting malgo device: %w", err)
	}

	return player, nil
}

type Player struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device

	playbackMU sync.Mutex
	playbacks  map[*Playback]struct{}
}

func (p *Player) CreateMedia(info audio.MediaInfo) *Media {
	decoder, err := mp3.NewDecoder(bytes.NewReader(info.Data))
	if err != nil {
		log.Error("Error creating decoder: %v", err)
		return &Media{}
	}

	if decoder.SampleRate() != 44100 {
		log.Warn("Unsupported sample rate: %d", decoder.SampleRate())
		return &Media{}
	}

	data, err := io.ReadAll(decoder)
	if err != nil {
		log.Error("Error reading decoder: %v", err)
		return &Media{}
	}
	buffer := gblob.LittleEndianBlock(data)

	frames := make([]mediaSample, len(data)/4)
	for i := 0; i < len(frames); i++ {
		frames[i].Left = buffer.Int16(i*4 + 0)
		frames[i].Right = buffer.Int16(i*4 + 2)
	}
	return &Media{
		frames: frames,
	}
}

func (p *Player) Play(media *Media, info audio.PlayInfo) *Playback {
	p.playbackMU.Lock()
	defer p.playbackMU.Unlock()

	playback := &Playback{
		media:  media,
		loop:   info.Loop,
		offset: 0,
	}
	p.playbacks[playback] = struct{}{}
	return playback
}

func (p *Player) Close() {
	p.device.Stop()
	p.device.Uninit()
	p.ctx.Uninit()
	p.ctx.Free()
}

func (p *Player) onSamples(pOutputSample, pInputSamples []byte, framecount uint32) {
	p.playbackMU.Lock()
	defer p.playbackMU.Unlock()

	buffer := gblob.LittleEndianBlock(pOutputSample)

	for i := 0; i < int(framecount); i++ {
		var left int64
		var right int64

		for playback := range p.playbacks {
			sample, ok := playback.Frame()
			if !ok {
				delete(p.playbacks, playback)
				continue
			}
			left += int64(sample.Left)
			right += int64(sample.Right)
		}

		left = min(int64(math.MaxInt16), left)
		right = min(int64(math.MaxInt16), right)
		left = max(int64(math.MinInt16), left)
		right = max(int64(math.MinInt16), right)

		buffer.SetInt16(i*4+0, int16(left))
		buffer.SetInt16(i*4+2, int16(right))
	}
}

func (p *Player) onStop() {
	p.playbackMU.Lock()
	defer p.playbackMU.Unlock()
	clear(p.playbacks)
}
