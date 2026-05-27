package audio

import (
	"fmt"

	"github.com/mokiat/lacking-native/audio/internal"
	"github.com/mokiat/lacking/audio"
)

func NewAPI() (*API, error) {
	graph := internal.NewGraph()

	player, err := internal.NewPlayer(graph)
	if err != nil {
		return nil, fmt.Errorf("error creating player: %w", err)
	}

	return &API{
		graph:  graph,
		player: player,
	}, nil
}

var _ audio.API = (*API)(nil)

type API struct {
	graph  *internal.Graph
	player *internal.Player
}

func (a *API) SampleRate() int {
	return a.player.SampleRate()
}

func (a *API) CreateMedia(data audio.MediaData) audio.Media {
	return a.player.CreateMedia(data)
}

func (a *API) Output() audio.Node {
	return a.player.Output()
}

func (a *API) SpatialListener() audio.SpatialListener {
	return a.player.SpatialListener()
}

func (a *API) CreatePlaybackNode(media audio.Media) audio.PlaybackNode {
	return a.player.CreatePlaybackNode(media.(*internal.Media))
}

func (a *API) CreateOscillatorNode() audio.OscillatorNode {
	return a.player.CreateOscillatorNode()
}

func (a *API) CreateGainNode() audio.GainNode {
	return a.player.CreateGainNode()
}

func (a *API) CreatePanNode() audio.PanNode {
	return a.player.CreatePanNode()
}

func (a *API) CreateSpatialNode() audio.SpatialNode {
	return a.player.CreateSpatialNode()
}

func (a *API) CreateHighPassNode() audio.HighPassNode {
	return a.player.CreateHighPassNode()
}

func (a *API) CreateLowPassNode() audio.LowPassNode {
	return a.player.CreateLowPassNode()
}

func (a *API) CreateDelayNode() audio.DelayNode {
	return a.player.CreateDelayNode()
}

func (a *API) CreateReverbNode() audio.ReverbNode {
	return a.player.CreateReverbNode()
}

func (a *API) CreateCompressorNode() audio.CompressorNode {
	return a.player.CreateCompressorNode()
}

func (a *API) CreateConnectorNode() audio.ConnectorNode {
	return a.player.CreateConnectorNode()
}

func (a *API) Chain(nodes ...audio.Node) {
	count := len(nodes)
	for i := 1; i < count; i++ {
		a.Connect(nodes[i-1], nodes[i])
	}
}

func (a *API) Connect(source, target audio.Node) {
	a.graph.Connect(source.(internal.Node), target.(internal.Node))
}

func (a *API) Disconnect(source, target audio.Node) {
	a.graph.Disconnect(source.(internal.Node), target.(internal.Node))
}

func (a *API) Play(media audio.Media, info audio.PlayInfo) audio.Playback {
	return a.player.Play(media.(*internal.Media), info)
}

func (a *API) Close() {
	a.player.Close()
}
