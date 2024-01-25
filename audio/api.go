package audio

import (
	"fmt"

	"github.com/mokiat/lacking-native/audio/internal"
	"github.com/mokiat/lacking/audio"
)

func NewAPI() (*API, error) {
	player, err := internal.NewPlayer()
	if err != nil {
		return nil, fmt.Errorf("error creating player: %w", err)
	}
	return &API{
		player: player,
	}, nil
}

var _ audio.API = (*API)(nil)

type API struct {
	player *internal.Player
}

func (a *API) CreateMedia(info audio.MediaInfo) audio.Media {
	return a.player.CreateMedia(info)
}

func (a *API) Play(media audio.Media, info audio.PlayInfo) audio.Playback {
	return a.player.Play(media.(*internal.Media), info)
}

func (a *API) Close() {
	a.player.Close()
}
