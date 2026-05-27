package internal

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/mokiat/lacking/audio"
)

const defaultSampleRate = 44100

func NewPlayer(graph *Graph) (*Player, error) {
	output := newOutputNode()
	graph.Register(output, true)

	player := &Player{
		playbacks: make(map[*Playback]struct{}),

		ioBuffer: newBuffer(1024 * 1024), // 1 MiB buffer
		listener: NewSpatialListener(),
		graph:    graph,
		output:   output,
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating malgo context: %w", err)
	}
	player.ctx = ctx

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = defaultSampleRate
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

	ioBuffer *Buffer
	graph    *Graph
	listener *SpatialListener
	output   *OutputNode

	inputCache []FrameList
}

func (p *Player) SampleRate() int {
	return defaultSampleRate
}

func (p *Player) CreateMedia(data audio.MediaData) *Media {
	frames := data.Frames
	if data.SampleRate != p.SampleRate() {
		logger.Warn("Resampling media",
			slog.Int("from", data.SampleRate),
			slog.Int("to", p.SampleRate()),
		)
		frames = audio.Resample(data.Frames, data.SampleRate, p.SampleRate())
	}
	return &Media{
		sampleRate: p.SampleRate(),
		frames:     frames,
	}
}

func (p *Player) Output() *OutputNode {
	return p.output
}

func (p *Player) SpatialListener() *SpatialListener {
	return p.listener
}

func (p *Player) CreatePlaybackNode(media *Media) *PlaybackNode {
	result := NewPlaybackNode(p, media)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeletePlaybackNode(node *PlaybackNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateOscillatorNode() *OscillatorNode {
	result := NewOscillatorNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteOscillatorNode(node *OscillatorNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateGainNode() *GainNode {
	result := NewGainNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteGainNode(node *GainNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreatePanNode() *PanNode {
	result := NewPanNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeletePanNode(node *PanNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateSpatialNode() *SpatialNode {
	result := NewSpatialNode(p, p.listener)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteSpatialNode(node *SpatialNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateHighPassNode() *HighPassNode {
	result := NewHighPassNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteHighPassNode(node *HighPassNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateLowPassNode() *LowPassNode {
	result := NewLowPassNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteLowPassNode(node *LowPassNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateDelayNode() *DelayNode {
	result := NewDelayNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteDelayNode(node *DelayNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateReverbNode() *ReverbNode {
	result := NewReverbNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteReverbNode(node *ReverbNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateCompressorNode() *CompressorNode {
	result := NewCompressorNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteCompressorNode(node *CompressorNode) {
	p.graph.Unregister(node)
}

func (p *Player) CreateConnectorNode() *ConnectorNode {
	result := NewConnectorNode(p)
	p.graph.Register(result, false)
	return result
}

func (p *Player) DeleteConnectorNode(node *ConnectorNode) {
	p.graph.Unregister(node)
}

func (p *Player) Play(media *Media, info audio.PlayInfo) *Playback {
	srcNode := p.CreatePlaybackNode(media)
	srcNode.SetLoop(info.Loop)
	srcNode.Start(0.0)
	panNode := p.CreatePanNode()
	panNode.SetPan(float32(info.Pan))
	gainNode := p.CreateGainNode()
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

	p.trackPlayback(playback)
	return playback
}

func (p *Player) trackPlayback(playback *Playback) {
	// Delete finished playbacks.
	for existing := range p.playbacks {
		if !existing.srcNode.IsPlaying() {
			p.deletePlayback(existing)
		}
	}

	// Add new playback.
	p.playbacks[playback] = struct{}{}
}

func (p *Player) onSamples(outputData, _ []byte, frameCount uint32) {
	clear(outputData)
	p.ioBuffer.Reset(frameCount)
	p.output.Prepare(outputData)

	snapshot := p.graph.Snapshot()
	p.processSnapshot(ProcessContext{
		SampleRate: defaultSampleRate,
		FrameCount: frameCount,
	}, snapshot)
}

func (p *Player) processSnapshot(ctx ProcessContext, snapshot *ProcessingSnapshot) {
	p.inputCache = p.inputCache[:0]
	for range len(snapshot.Processings) {
		p.inputCache = append(p.inputCache, p.ioBuffer.Allocate())
	}
	outputCache := p.ioBuffer.Allocate()

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

func (p *Player) Close() {
	p.device.Stop()
	p.device.Uninit()
	p.ctx.Uninit()
	p.ctx.Free()
}

func (p *Player) onStop() {
	p.playbackMU.Lock()
	defer p.playbackMU.Unlock()

	for playback := range p.playbacks {
		p.deletePlayback(playback)
	}
}

func (p *Player) deletePlayback(playback *Playback) {
	playback.release()
	delete(p.playbacks, playback)
}
