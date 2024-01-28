package internal

import (
	"time"

	"github.com/mokiat/gomath/sprec"
)

type Media struct {
	sampleRate   int
	length       int
	leftChannel  Channel
	rightChannel Channel
}

func (m *Media) Length() time.Duration {
	if m == nil {
		return 0
	}
	return time.Second * time.Duration(m.length/m.sampleRate)
}

func (m *Media) Delete() {
	if m == nil {
		return
	}
	m.length = 0
	m.sampleRate = 1
	m.leftChannel.samples = nil
	m.rightChannel.samples = nil
}

type Channel struct {
	samples []float32
}

type MediaFrame struct {
	Left  float32
	Right float32
}

func (f *MediaFrame) Add(other MediaFrame) {
	f.Left += other.Left
	f.Right += other.Right
}

func (f *MediaFrame) ApplyGain(gain float32) {
	f.Left *= gain
	f.Right *= gain
}

func (f *MediaFrame) ApplyPan(pan float32) {
	// Note: This algorithm is consistent with WebAudio's pan algorithm.
	// https://webaudio.github.io/web-audio-api/#stereopanner-algorithm
	pan = max(min(pan, 1.0), -1.0)

	angleFraction := pan
	if angleFraction < 0.0 {
		angleFraction += 1.0
	}

	leftGain := sprec.Cos(sprec.Radians(angleFraction * sprec.Pi / 2.0))
	rightGain := sprec.Sin(sprec.Radians(angleFraction * sprec.Pi / 2.0))

	var newLeft, newRight float32
	if pan >= 0.0 {
		newLeft = (f.Left * leftGain)
		newRight = (f.Left * rightGain) + f.Right
	} else {
		newLeft = (f.Right * leftGain) + f.Left
		newRight = (f.Right * rightGain)
	}
	f.Left = newLeft
	f.Right = newRight
}

func (f *MediaFrame) Clamp() {
	f.Left = max(min(f.Left, 1.0), -1.0)
	f.Right = max(min(f.Right, 1.0), -1.0)
}
