package internal

import (
	"bytes"
	"fmt"
	"io"
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
		return nil
	}

	if decoder.SampleRate() != 44100 {
		//  TODO: Handle resample in the future.
		log.Error("Unsupported sample rate: %d", decoder.SampleRate())
		return nil
	}

	data, err := io.ReadAll(decoder)
	if err != nil {
		log.Error("Error reading decoder: %v", err)
		return nil
	}
	buffer := gblob.LittleEndianBlock(data)

	length := len(data) / 4
	leftChannel := Channel{
		samples: make([]float32, length),
	}
	rightChannel := Channel{
		samples: make([]float32, length),
	}
	for i := 0; i < length; i++ {
		leftInt16 := buffer.Int16(i*4 + 0)
		rightInt16 := buffer.Int16(i*4 + 2)
		leftChannel.samples[i] = int16ToFloat32(leftInt16)
		rightChannel.samples[i] = int16ToFloat32(rightInt16)
	}

	return &Media{
		sampleRate:   44100,
		length:       length,
		leftChannel:  leftChannel,
		rightChannel: rightChannel,
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
		var aggFrame MediaFrame

		for playback := range p.playbacks {
			frame, ok := playback.Frame()
			if !ok {
				delete(p.playbacks, playback)
				continue
			}
			aggFrame.Add(frame)
		}

		aggFrame.Clamp()
		buffer.SetInt16(i*4+0, float32ToInt16(aggFrame.Left))
		buffer.SetInt16(i*4+2, float32ToInt16(aggFrame.Right))
	}
}

func (p *Player) onStop() {
	p.playbackMU.Lock()
	defer p.playbackMU.Unlock()
	clear(p.playbacks)
}
