package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewSpatialListener() *SpatialListener {
	return &SpatialListener{
		position: sprec.ZeroVec3(),
		rotation: sprec.IdentityQuat(),
	}
}

type SpatialListener struct {
	mu       sync.Mutex
	position sprec.Vec3
	rotation sprec.Quat
}

var _ audio.SpatialListener = (*SpatialListener)(nil)

func (l *SpatialListener) Position() sprec.Vec3 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.position
}

func (l *SpatialListener) SetPosition(position sprec.Vec3) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.position = position
}

func (l *SpatialListener) Rotation() sprec.Quat {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotation
}

func (l *SpatialListener) SetRotation(rotation sprec.Quat) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rotation = rotation
}
