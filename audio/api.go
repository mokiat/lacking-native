package audio

import (
	"fmt"

	"github.com/mokiat/gog/ds"
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

	gainPool *ds.Stack[*internal.GainNode]
}

func (a *API) CreateMedia(info audio.MediaInfo) audio.Media {
	return a.player.CreateMedia(info)
}

func (a *API) Play(media audio.Media, info audio.PlayInfo) audio.Playback {
	return a.player.Play(media.(*internal.Media), info)
}

func (a *API) CreatePlayback(media audio.Media, loop bool) audio.PlaybackNode {
	return a.player.CreatePlayback(media.(*internal.Media), loop)
}

func (a *API) CreateOscillator() audio.OscillatorNode {
	return a.player.CreateOscillator()
}

func (a *API) CreateGain() audio.GainNode {
	return a.player.CreateGain()
}

func (a *API) CreatePan() audio.PanNode {
	return a.player.CreatePan()
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

func (a *API) Output() audio.Node {
	return a.player.Output()
}

func (a *API) Close() {
	a.player.Close()
}
