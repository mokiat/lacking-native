package audio

import (
	"fmt"
	"log/slog"
	"math"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/mokiat/gblob"
	"github.com/mokiat/lacking-native/core/audio/internal"
	"github.com/mokiat/lacking/core/audio"
)

type API struct {
	ctx        *malgo.AllocatedContext
	device     *malgo.Device
	sampleRate int

	masterBus *internal.MasterBus

	ioBuffer  *internal.Buffer
	unitInput [][]audio.Frame

	pipelineMu         sync.Mutex
	activePipeline     *internal.Pipeline
	pendingPipeline    *internal.Pipeline
	tempPipeline       *internal.Pipeline
	mustUpdatePipeline bool
}

var _ audio.API = (*API)(nil)

func NewAPI() (*API, error) {
	api := &API{
		ioBuffer:  internal.NewBuffer(128 * 1024),
		unitInput: make([][]audio.Frame, 0, 128),

		masterBus: internal.NewMasterBus(),

		activePipeline:  internal.NewPipeline(128),
		pendingPipeline: internal.NewPipeline(128),
		tempPipeline:    internal.NewPipeline(128),
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("init malgo context: %w", err)
	}
	api.ctx = ctx

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.Alsa.NoMMap = 1

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: api.onData,
		Stop: api.onStop,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return nil, fmt.Errorf("init malgo device: %w", err)
	}
	api.device = device
	api.sampleRate = int(device.SampleRate())

	// NOTE: By this time the API must be able to handle callbacks, since Start
	// may trigger onSamples immediately.

	if err := device.Start(); err != nil {
		return nil, fmt.Errorf("error starting malgo device: %w", err)
	}

	return api, nil
}

// Destroy cleans up resources used by the API. It should be called when the API is no longer needed.
func (a *API) Destroy() {
	if err := a.device.Stop(); err != nil {
		logger.Error("Error stopping malgo device",
			slog.String("error", err.Error()),
		)
	}
	a.device.Uninit()

	if err := a.ctx.Uninit(); err != nil {
		logger.Error("Error uninitializing malgo context",
			slog.String("error", err.Error()),
		)
	}
	a.ctx.Free()
}

func (a *API) CreateMedia(data audio.MediaData) audio.Media {
	return internal.NewMedia(data)
}

func (a *API) CreateBus(settings audio.BusSettings) audio.Bus {
	panic("TODO: implement CreateBus")
}

func (a *API) CreatePlayback(bus audio.Bus, media audio.Media, settings audio.PlaybackSettings) audio.Playback {
	panic("TODO")
}

func (a *API) CreateSpatialPlayback(bus audio.Bus, media audio.Media, settings audio.PlaybackSettings) audio.SpatialPlayback {
	panic("TODO")
}

func (a *API) MasterBus() audio.MasterBus {
	return a.masterBus
}

func (a *API) SpatialListener() audio.SpatialListener {
	panic("TODO")
}

func (a *API) onData(outputData, _ []byte, frameCount uint32) {
	pipeline := a.pickPipeline()

	a.ioBuffer.Reset()

	ctx := internal.ProcessContext{
		SampleRate: int(a.device.SampleRate()),
		FrameCount: int(frameCount),
		Buffer:     a.ioBuffer,
	}
	outputFrames := a.processPipeline(ctx, pipeline)

	buffer := gblob.LittleEndianBlock(outputData)
	for i, frame := range outputFrames {
		buffer.SetInt16(i*4+0, float32ToInt16(frame.Left))
		buffer.SetInt16(i*4+2, float32ToInt16(frame.Right))
	}
}

func (a *API) pickPipeline() *internal.Pipeline {
	a.pipelineMu.Lock()
	defer a.pipelineMu.Unlock()
	if a.mustUpdatePipeline {
		a.mustUpdatePipeline = false
		a.activePipeline, a.pendingPipeline = a.pendingPipeline, a.activePipeline
	}
	return a.activePipeline
}

func (a *API) processPipeline(ctx internal.ProcessContext, pipeline *internal.Pipeline) []audio.Frame {
	a.unitInput = a.unitInput[:0]
	defer clear(a.unitInput) // clear refs
	for range len(pipeline.Units) {
		a.unitInput = append(a.unitInput, a.ioBuffer.Allocate(ctx.FrameCount))
	}

	outputFrames := a.ioBuffer.Allocate(ctx.FrameCount)
	for i, unit := range pipeline.Units {
		unitOutput := unit.Processor.Process(ctx, a.unitInput[i])
		if unit.TargetIndex >= 0 {
			applyFrames(a.unitInput[unit.TargetIndex], unitOutput)
		} else {
			applyFrames(outputFrames, unitOutput)
		}
	}

	return outputFrames
}

func (a *API) onStop() {
	// Unclear what to do here. Some documents indicate that this can be called
	// by the OS when the audio device is stopped, but this has not yet
	// been observed.
}

func float32ToInt16(value float32) int16 {
	value = max(-1.0, min(value, 1.0)) // prevent overflow
	if value >= 0.0 {
		return int16(value * float32(math.MaxInt16))
	} else {
		return int16(-value * float32(math.MinInt16))
	}
}

func applyFrames(target, source []audio.Frame) {
	for i := range target {
		target[i].Left += source[i].Left
		target[i].Right += source[i].Right
	}
}
