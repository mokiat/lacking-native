package audio

import (
	"github.com/mokiat/lacking-native/audio/internal"
	"github.com/mokiat/lacking/audio"
)

func NewAPI() *API {
	return &API{
		player: internal.NewPlayer(),
	}
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
