package internal

import (
	"time"
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

func (f *MediaFrame) Clamp() {
	f.Left = max(min(f.Left, 1.0), -1.0)
	f.Right = max(min(f.Right, 1.0), -1.0)
}
