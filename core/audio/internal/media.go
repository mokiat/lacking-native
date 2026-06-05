package internal

import "github.com/mokiat/lacking/core/audio"

// Media holds decoded audio frames and their sample rate.
type Media struct {
	sampleRate int
	frames     []audio.Frame
}

var _ audio.Media = (*Media)(nil)

// NewMedia creates a Media from decoded audio data.
func NewMedia(data audio.MediaData) *Media {
	return &Media{
		sampleRate: data.SampleRate,
		frames:     data.Frames,
	}
}

// Length returns the duration of the media in seconds.
func (m *Media) Length() float32 {
	frameCount := len(m.frames)
	return audio.Seconds(frameCount, m.sampleRate)
}

// Release frees the frame data. Existing playbacks are not affected.
func (m *Media) Release() {
	m.frames = nil
}
