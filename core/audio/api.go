package audio

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/mokiat/gblob"
	"github.com/mokiat/lacking-native/core/audio/internal"
	"github.com/mokiat/lacking/core/audio"
)

// API implements the audio.API interface using the miniaudio (malgo) backend.
// It manages the audio device, maintains a triple-buffered processing pipeline,
// and owns the master bus and spatial listener.
type API struct {
	ctx        *malgo.AllocatedContext
	device     *malgo.Device
	sampleRate int

	masterBus *internal.MasterBus
	listener  *internal.SpatialListener

	ioBuffer  *internal.Buffer
	unitInput [][]audio.Frame

	pipelineMu         sync.Mutex
	activePipeline     *internal.Pipeline
	pendingPipeline    *internal.Pipeline
	tempPipeline       *internal.Pipeline
	mustUpdatePipeline bool
}

var _ audio.API = (*API)(nil)

// NewAPI initializes a new API backed by the default miniaudio playback
// device at 44100 Hz stereo 16-bit output. The audio device starts immediately
// upon successful return.
func NewAPI() (*API, error) {
	api := &API{
		ioBuffer:  internal.NewBuffer(128 * 1024),
		unitInput: make([][]audio.Frame, 0, 128),

		masterBus: internal.NewMasterBus(),
		listener:  internal.NewSpatialListener(),

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

// CreateMedia creates a new media object from the provided media data.
func (a *API) CreateMedia(data audio.MediaData) audio.Media {
	return internal.NewMedia(data)
}

// CreateBus creates a new audio bus registered with the master bus.
func (a *API) CreateBus(settings audio.BusSettings) audio.Bus {
	return internal.NewBus(
		a.masterBus,
		settings,
		a.sampleRate,
		a.invalidate,
	)
}

// CreatePlayback creates a non-spatial playback on the given bus.
func (a *API) CreatePlayback(bus audio.Bus, media audio.Media, settings audio.PlaybackSettings) audio.Playback {
	base := internal.NewBasePlayback(
		bus.(*internal.Bus),
		media.(*internal.Media),
		settings,
		a.sampleRate,
	)
	return internal.NewDefaultPlayback(base)
}

// CreateSpatialPlayback creates a spatially positioned playback on the given
// bus, attached to the API's shared spatial listener.
func (a *API) CreateSpatialPlayback(bus audio.Bus, media audio.Media, settings audio.PlaybackSettings) audio.SpatialPlayback {
	base := internal.NewBasePlayback(
		bus.(*internal.Bus),
		media.(*internal.Media),
		settings,
		a.sampleRate,
	)
	return internal.NewSpatialPlayback(base, a.listener)
}

// MasterBus returns the master bus for the audio system.
func (a *API) MasterBus() audio.MasterBus {
	return a.masterBus
}

// SpatialListener returns the spatial listener used for 3D audio positioning.
func (a *API) SpatialListener() audio.SpatialListener {
	return a.listener
}

// onData is the miniaudio callback invoked on the real-time audio thread to
// fill the output buffer.
func (a *API) onData(outputData, _ []byte, frameCount uint32) {
	pipeline := a.pickPipeline()

	a.ioBuffer.Reset()

	ctx := internal.ProcessContext{
		SampleRate: a.sampleRate,
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

// pickPipeline swaps in the pending pipeline if one is available and returns
// the active pipeline. Called from the real-time audio thread.
func (a *API) pickPipeline() *internal.Pipeline {
	a.pipelineMu.Lock()
	defer a.pipelineMu.Unlock()
	if a.mustUpdatePipeline {
		a.mustUpdatePipeline = false
		a.activePipeline, a.pendingPipeline = a.pendingPipeline, a.activePipeline
	}
	return a.activePipeline
}

// processPipeline runs every unit in the pipeline in topological order,
// accumulating outputs into their target units' input buffers, and returns the
// final mixed output. Called from the real-time audio thread.
func (a *API) processPipeline(ctx internal.ProcessContext, pipeline *internal.Pipeline) []audio.Frame {
	a.unitInput = a.unitInput[:0]
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

	clear(a.unitInput) // clear refs

	return outputFrames
}

// invalidate rebuilds the processing pipeline on the game thread and makes it
// available to the audio thread via the triple-buffer swap.
func (a *API) invalidate() {
	a.constructPipeline(a.tempPipeline)

	a.pipelineMu.Lock()
	defer a.pipelineMu.Unlock()
	a.mustUpdatePipeline = true
	a.tempPipeline, a.pendingPipeline = a.pendingPipeline, a.tempPipeline
}

// constructPipeline populates target with units ordered so that each
// processor's inputs are fully accumulated before it runs (playbacks →
// buses → master). Units are built in reverse order then flipped, with
// TargetIndex values adjusted accordingly.
func (a *API) constructPipeline(target *internal.Pipeline) {
	clear(target.Units) // clear refs
	target.Units = target.Units[:0]

	target.Units = append(target.Units, internal.Unit{
		Processor:   a.masterBus,
		TargetIndex: -1, // send to output
	})

	for _, bus := range a.masterBus.Buses() {
		if !bus.IsPlaying() {
			continue
		}

		busUnitIndex := len(target.Units)
		target.Units = append(target.Units, internal.Unit{
			Processor:   bus,
			TargetIndex: 0, // send to master bus
		})

		for _, playback := range bus.Playbacks() {
			target.Units = append(target.Units, internal.Unit{
				Processor:   playback,
				TargetIndex: busUnitIndex, // send to bus
			})
		}
	}

	count := len(target.Units)
	for i := range count {
		unit := &target.Units[i]
		if unit.TargetIndex >= 0 {
			unit.TargetIndex = count - unit.TargetIndex - 1
		}
	}
	slices.Reverse(target.Units)
}

func (a *API) onStop() {
	// Unclear what to do here. Some documents indicate that this can be called
	// by the OS when the audio device is stopped, but this has not yet
	// been observed.
}

// float32ToInt16 converts a normalized float32 sample in [-1, 1] to int16,
// using the full range of both positive and negative int16 values.
func float32ToInt16(value float32) int16 {
	value = max(-1.0, min(value, 1.0)) // prevent overflow
	if value >= 0.0 {
		return int16(value * float32(math.MaxInt16))
	} else {
		return int16(-value * float32(math.MinInt16))
	}
}

// applyFrames accumulates source into target by adding each channel sample.
func applyFrames(target, source []audio.Frame) {
	for i := range target {
		target[i].Left += source[i].Left
		target[i].Right += source[i].Right
	}
}
