package internal

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/hajimehoshi/go-mp3"
	"github.com/mokiat/gblob"
	"github.com/mokiat/lacking/audio"
)

const sampleRate = 44100

func NewPlayer(graph *Graph) (*Player, error) {
	output := newOutputNode()
	graph.Register(output)

	player := &Player{
		playbacks: make(map[*Playback]struct{}),
		output:    output,
		buffer:    newBuffer(1024 * 1024), // 1 MiB buffer
		graph:     graph,
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating malgo context: %w", err)
	}
	player.ctx = ctx

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = sampleRate
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

	buffer *Buffer
	graph  *Graph
	output *OutputNode

	inputCache []FrameList
}

func (p *Player) CreateMedia(info audio.MediaInfo) *Media {
	decoder, err := mp3.NewDecoder(bytes.NewReader(info.Data))
	if err != nil {
		logger.Error("Error creating decoder",
			slog.String("error", err.Error()),
		)
		return nil
	}

	if decoder.SampleRate() != 44100 {
		//  TODO: Handle resample in the future.
		logger.Error("Unsupported sample rate",
			slog.Int("rate", decoder.SampleRate()),
		)
		return nil
	}

	data, err := io.ReadAll(decoder)
	if err != nil {
		logger.Error("Error reading decoder",
			slog.String("error", err.Error()),
		)
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
	for i := range length {
		leftInt16 := buffer.Int16(i*4 + 0)
		rightInt16 := buffer.Int16(i*4 + 2)
		leftChannel.samples[i] = int16ToFloat32(leftInt16)
		rightChannel.samples[i] = int16ToFloat32(rightInt16)
	}

	return &Media{
		sampleRate:   44100,
		length:       uint64(length),
		leftChannel:  leftChannel,
		rightChannel: rightChannel,
	}
}

func (p *Player) Play(media *Media, info audio.PlayInfo) *Playback {
	srcNode := p.CreatePlayback(media, info.Loop)
	panNode := p.CreatePan()
	panNode.SetPan(float32(info.Pan))
	gainNode := p.CreateGain()
	gainNode.SetGain(float32(info.Gain.ValueOrDefault(1.0)))

	p.graph.Connect(srcNode, panNode)
	p.graph.Connect(panNode, gainNode)
	p.graph.Connect(gainNode, p.output)

	playback := &Playback{
		srcNode:  srcNode,
		panNode:  panNode,
		gainNode: gainNode,
	}

	p.playbackMU.Lock()
	defer p.playbackMU.Unlock()
	// Delete finished playbacks.
	for playback := range p.playbacks {
		if playback.srcNode.Done() {
			p.deletePlayback(playback)
		}
	}
	// Add new playback.
	p.playbacks[playback] = struct{}{}

	return playback
}

func (p *Player) Close() {
	p.device.Stop()
	p.device.Uninit()
	p.ctx.Uninit()
	p.ctx.Free()
}

func (p *Player) CreatePlayback(media *Media, loop bool) *PlaybackNode {
	// TODO: Fetch from pool.
	result := NewPlaybackNode(p, media, loop)
	p.graph.Register(result)
	return result
}

func (p *Player) DeletePlayback(node *PlaybackNode) {
	p.graph.Unregister(node)
	// TODO: Return to pool.
}

func (p *Player) CreateOscillator() *OscillatorNode {
	// TODO: Fetch from pool.
	result := NewOscillatorNode(p)
	p.graph.Register(result)
	return result
}

func (p *Player) DeleteOscillator(node *OscillatorNode) {
	p.graph.Unregister(node)
	// TODO: Return to pool.
}

func (p *Player) CreateGain() *GainNode {
	// TODO: Fetch from pool.
	result := NewGainNode(p)
	p.graph.Register(result)
	return result
}

func (p *Player) DeleteGain(node *GainNode) {
	p.graph.Unregister(node)
	// TODO: Return to pool.
}

func (p *Player) CreatePan() *PanNode {
	// TODO: Fetch from pool.
	result := NewPanNode(p)
	p.graph.Register(result)
	return result
}

func (p *Player) DeletePan(node *PanNode) {
	p.graph.Unregister(node)
	// TODO: Return to pool.
}

func (p *Player) Output() *OutputNode {
	return p.output
}

func (p *Player) onSamples(outputData, _ []byte, frameCount uint32) {
	clear(outputData)
	p.buffer.Reset(frameCount)
	p.output.Prepare(outputData)

	snapshot := p.graph.Snapshot()
	p.processSnapshot(ProcessContext{
		SampleRate: sampleRate,
		FrameCount: frameCount,
	}, snapshot)
}

func (p *Player) processSnapshot(ctx ProcessContext, snapshot *ProcessingSnapshot) {
	p.inputCache = p.inputCache[:0]
	for range len(snapshot.Processings) {
		p.inputCache = append(p.inputCache, p.buffer.Allocate())
	}
	outputCache := p.buffer.Allocate()

	for sourceIndex, processing := range snapshot.Processings {
		if targetIndex, ok := snapshot.IsDirectConnection(uint32(sourceIndex)); ok {
			// This is an optimization that passes the input of the next processing
			// unit directly as an output to the current one, avoiding an extra copy.
			processing.Processor.Process(ctx, p.inputCache[sourceIndex], p.inputCache[targetIndex])
		} else {
			// General processing path that handles multiple output assignments.
			// In such cases the output is first written to a temporary cache and then
			// distributed to the target processing units.
			// This is also necessary if the target processing unit has multiple
			// inputs in order to avoid overwriting data.
			clear(outputCache)
			processing.Processor.Process(ctx, p.inputCache[sourceIndex], outputCache)
			assignmentIndex := processing.AssignmentOffset
			for range processing.OutputCount {
				targetIndex := snapshot.Assignments[assignmentIndex]
				p.inputCache[targetIndex].Add(outputCache)
				assignmentIndex++
			}
		}
	}
}

func (p *Player) onStop() {
	p.playbackMU.Lock()
	defer p.playbackMU.Unlock()

	for playback := range p.playbacks {
		p.deletePlayback(playback)
	}
}

func (p *Player) deletePlayback(playback *Playback) {
	playback.srcNode.Delete()
	playback.panNode.Delete()
	playback.gainNode.Delete()
	delete(p.playbacks, playback)
}
