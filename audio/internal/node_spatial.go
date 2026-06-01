package internal

import (
	"sync"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func newSpatialNode(player *Player, listener *SpatialListener) *SpatialNode {
	return &SpatialNode{
		player:   player,
		listener: listener,
	}
}

type SpatialNode struct {
	audio.Node // marker interface

	player   *Player
	listener *SpatialListener

	mu       sync.Mutex
	position sprec.Vec3
}

var _ Node = (*SpatialNode)(nil)
var _ audio.SpatialNode = (*SpatialNode)(nil)

func (n *SpatialNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	position := n.Position()                  // store position locally to avoid long locks
	listenerPosition := n.listener.Position() // store listener position locally to avoid long locks
	listenerRotation := n.listener.Rotation() // store listener rotation locally to avoid long locks

	// This implementation is consistent with WebAudio's equal power panning algorithm.
	// https://webaudio.github.io/web-audio-api/#Spatialization-equal-power-panning

	// This works by treating the pan value as an angle on a semicircle.
	// Furthermore, it always keeps one channel fully audible, while the other
	// channel is attenuated and blended based on the pan position.
	// This is a bit different from standard mono panning approaches.

	azimuth, _, distance := n.calculateAzimuthElevationDistance(position, listenerPosition, listenerRotation)

	// wrap azimuth to [-pi/2, pi/2]
	piAngle := sprec.Radians(sprec.Pi)
	halfPiAngle := piAngle / 2.0
	switch {
	case azimuth < -halfPiAngle:
		azimuth = -piAngle - azimuth
	case azimuth > halfPiAngle:
		azimuth = piAngle - azimuth
	}

	panAngle := gog.Ternary(azimuth >= 0.0, azimuth, azimuth+halfPiAngle)
	leftGain := sprec.Cos(panAngle)
	rightGain := sprec.Sin(panAngle)

	distanceGain := 1.0 / max(1.0, distance)

	if azimuth < 0.0 {
		for i, frame := range inputFrames {
			outputFrames[i] = Frame{
				Left:  distanceGain * ((frame.Right * leftGain) + frame.Left),
				Right: distanceGain * (frame.Right * rightGain),
			}
		}
	} else {
		for i, frame := range inputFrames {
			outputFrames[i] = Frame{
				Left:  distanceGain * (frame.Left * leftGain),
				Right: distanceGain * ((frame.Left * rightGain) + frame.Right),
			}
		}
	}
}

func (n *SpatialNode) Position() sprec.Vec3 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.position
}

func (n *SpatialNode) SetPosition(position sprec.Vec3) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.position = position
}

func (n *SpatialNode) Delete() {
	n.player.DeleteSpatialNode(n)
}

func (n *SpatialNode) calculateAzimuthElevationDistance(emitterPosition, listenerPosition sprec.Vec3, listenerRotation sprec.Quat) (sprec.Angle, sprec.Angle, float32) {
	deltaPosition := sprec.Vec3Diff(emitterPosition, listenerPosition)
	deltaPosition = sprec.QuatVec3Rotation(
		sprec.InverseQuat(listenerRotation),
		deltaPosition,
	)

	distance := deltaPosition.Length()
	if distance < 0.0001 {
		return 0.0, 0.0, 0.0
	}

	deltaPosition = sprec.Vec3Quot(deltaPosition, distance)
	azimuth := sprec.NormalizeAngle(
		sprec.Atan2(deltaPosition.X, -deltaPosition.Z),
	)
	elevation := sprec.NormalizeAngle(
		sprec.Atan2(deltaPosition.Y, sprec.Sqrt(deltaPosition.X*deltaPosition.X+deltaPosition.Z*deltaPosition.Z)),
	)
	return azimuth, elevation, distance
}
