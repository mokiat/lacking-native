package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// SpatialListener holds the position, orientation, and velocity of the audio
// listener in world space. All accessors are safe for concurrent use across
// the game thread and the real-time audio callback.
type SpatialListener struct {
	mu       sync.Mutex
	position sprec.Vec3
	rotation sprec.Quat
}

var _ audio.SpatialListener = (*SpatialListener)(nil)

// NewSpatialListener creates a SpatialListener at the origin with identity
// rotation and zero velocity.
func NewSpatialListener() *SpatialListener {
	return &SpatialListener{
		position: sprec.ZeroVec3(),
		rotation: sprec.IdentityQuat(),
	}
}

// Position returns the listener's position in world space.
func (l *SpatialListener) Position() sprec.Vec3 {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.position
}

// SetPosition sets the listener's position in world space.
func (l *SpatialListener) SetPosition(position sprec.Vec3) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.position = position
}

// Rotation returns the listener's orientation in world space.
func (l *SpatialListener) Rotation() sprec.Quat {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.rotation
}

// SetRotation sets the listener's orientation in world space.
func (l *SpatialListener) SetRotation(rotation sprec.Quat) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.rotation = rotation
}
