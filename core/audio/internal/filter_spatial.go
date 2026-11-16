package internal

import (
	"sync"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// SpatialFilter is a stereo Processor that applies 3D spatial audio
// processing to the input frames. It attenuates and pans audio based on the
// relative position and orientation of the emitter and the listener. All
// accessors are safe for concurrent use across the game thread and the
// real-time audio callback.
type SpatialFilter struct {
	listener *SpatialListener

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu             sync.Mutex
	position       sprec.Vec3
	rotation       sprec.Quat
	innerConeAngle sprec.Angle
	outerConeAngle sprec.Angle
	outerConeGain  float32
}

var _ audio.SpatialEmitter = (*SpatialFilter)(nil)
var _ Processor = (*SpatialFilter)(nil)

// NewSpatialFilter creates a SpatialFilter attached to the given listener,
// positioned at the origin with identity rotation and a 360-degree
// omnidirectional cone.
func NewSpatialFilter(listener *SpatialListener) *SpatialFilter {
	return &SpatialFilter{
		listener: listener,

		position: sprec.ZeroVec3(),
		rotation: sprec.IdentityQuat(),

		innerConeAngle: sprec.Degrees(360),
		outerConeAngle: sprec.Degrees(360),
		outerConeGain:  0.0,
	}
}

// Position returns the emitter's position in world space.
func (f *SpatialFilter) Position() sprec.Vec3 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.position
}

// SetPosition sets the emitter's position in world space.
func (f *SpatialFilter) SetPosition(position sprec.Vec3) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.position = position
}

// Rotation returns the emitter's orientation in world space. The emitter's
// forward (cone) direction is its +Z axis.
func (f *SpatialFilter) Rotation() sprec.Quat {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.rotation
}

// SetRotation sets the emitter's orientation in world space.
func (f *SpatialFilter) SetRotation(rotation sprec.Quat) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.rotation = rotation
}

// InnerConeAngle returns the inner cone angle of the emitter.
func (f *SpatialFilter) InnerConeAngle() sprec.Angle {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.innerConeAngle
}

// SetInnerConeAngle sets the inner cone angle of the emitter.
func (f *SpatialFilter) SetInnerConeAngle(angle sprec.Angle) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.innerConeAngle = angle
}

// OuterConeAngle returns the outer cone angle of the emitter.
func (f *SpatialFilter) OuterConeAngle() sprec.Angle {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.outerConeAngle
}

// SetOuterConeAngle sets the outer cone angle of the emitter.
func (f *SpatialFilter) SetOuterConeAngle(angle sprec.Angle) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.outerConeAngle = angle
}

// OuterConeGain returns the gain applied when the listener is outside the
// outer cone.
func (f *SpatialFilter) OuterConeGain() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.outerConeGain
}

// SetOuterConeGain sets the gain applied when the listener is outside the
// outer cone.
func (f *SpatialFilter) SetOuterConeGain(gain float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.outerConeGain = gain
}

// Process applies 3D spatial audio processing to inputFrames and returns the
// result.
func (f *SpatialFilter) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	f.mu.Lock()
	position := f.position             // store value locally to avoid long locks
	rotation := f.rotation             // store value locally to avoid long locks
	innerConeAngle := f.innerConeAngle // store value locally to avoid long locks
	outerConeAngle := f.outerConeAngle // store value locally to avoid long locks
	outerConeGain := f.outerConeGain   // store value locally to avoid long locks
	f.mu.Unlock()

	f.listener.mu.Lock()
	listenerPosition := f.listener.position // store listener position locally to avoid long locks
	listenerRotation := f.listener.rotation // store listener rotation locally to avoid long locks
	f.listener.mu.Unlock()

	listenerOffset := sprec.Vec3Diff(listenerPosition, position)
	listenerDistance := listenerOffset.Length()

	azimuth, elevation := calculateAzimuthElevation(
		listenerOffset,
		listenerDistance,
		listenerRotation,
	)

	leftGain, rightGain, isLeftSide := calculateAzimuthElevationGain(
		azimuth,
		elevation,
	)

	distanceGain := calculateDistanceGain(
		listenerDistance,
	)

	coneGain := calculateConeGain(
		rotation.OrientationZ(),
		listenerOffset,
		listenerDistance,
		innerConeAngle,
		outerConeAngle,
		outerConeGain,
	)

	sampleGain := coneGain * distanceGain

	outputFrames := ctx.Buffer.Allocate(ctx.FrameCount)

	// NOTE: The side must come from the gain calculation and not from the
	// azimuth above, since the two can disagree once the azimuth is wrapped
	// (e.g. an emitter that is directly behind the listener).
	if isLeftSide {
		for i, frame := range inputFrames {
			outputFrames[i] = audio.Frame{
				Left:  sampleGain * ((frame.Right * leftGain) + frame.Left),
				Right: sampleGain * (frame.Right * rightGain),
			}
		}
	} else {
		for i, frame := range inputFrames {
			outputFrames[i] = audio.Frame{
				Left:  sampleGain * (frame.Left * leftGain),
				Right: sampleGain * ((frame.Left * rightGain) + frame.Right),
			}
		}
	}

	return outputFrames
}

// calculateAzimuthElevation returns the azimuth and elevation angles of the
// emitter relative to the listener's orientation, given the world-space offset
// from emitter to listener and its length.
func calculateAzimuthElevation(
	listenerOffset sprec.Vec3,
	distance float32,
	listenerRotation sprec.Quat,
) (sprec.Angle, sprec.Angle) {

	// If the distance is very small, we can treat the emitter as being at the
	// same position as the listener.
	if distance < 0.001 {
		return 0.0, 0.0
	}

	// Get the emitter position relative to the listener's orientation in
	// local coordinates.
	deltaPosition := sprec.QuatVec3Rotation(
		sprec.InverseQuat(listenerRotation),
		sprec.InverseVec3(listenerOffset), // from the point of view of the listener
	)

	emitterDirection := sprec.Vec3Quot(deltaPosition, distance)

	azimuth := sprec.Atan2(
		emitterDirection.X,
		-emitterDirection.Z,
	)
	elevation := sprec.Atan2(
		emitterDirection.Y,
		sprec.Sqrt(emitterDirection.X*emitterDirection.X+emitterDirection.Z*emitterDirection.Z),
	)

	return azimuth, elevation
}

// calculateAzimuthElevationGain returns the left and right channel gain
// factors for the given azimuth and elevation, using the WebAudio equal-power
// panning algorithm. The returned flag reports whether the emitter is
// considered to be on the left side of the listener, which determines how the
// gain factors are to be applied to the channels.
func calculateAzimuthElevationGain(
	azimuth sprec.Angle,
	elevation sprec.Angle,
) (float32, float32, bool) {

	// What follows is heavily based on WebAudio API's approach:
	// https://webaudio.github.io/web-audio-api/#Spatialization-equal-power-panning

	// wrap azimuth to [-pi/2, pi/2]
	piAngle := sprec.Radians(sprec.Pi)
	halfPiAngle := piAngle / 2.0
	switch {
	case azimuth < -halfPiAngle:
		azimuth = -piAngle - azimuth
	case azimuth > halfPiAngle:
		azimuth = piAngle - azimuth
	}

	// Apply elevation compensation to the azimuth angle to simulate the effect
	// of the sound source being above or below the listener. This is done
	// after the wrapping, as otherwise a source that is behind the listener
	// and above or below it would jump from one side to the other when it
	// crosses the center line.
	azimuth *= sprec.Angle(sprec.Cos(elevation))

	isLeftSide := azimuth < 0.0
	panAngle := gog.Ternary(isLeftSide, azimuth+halfPiAngle, azimuth)
	return sprec.Cos(panAngle), sprec.Sin(panAngle), isLeftSide
}

// calculateDistanceGain returns the gain factor for the given distance, using
// an inverse distance model clamped to a minimum distance of 1.
func calculateDistanceGain(distance float32) float32 {
	return 1.0 / max(1.0, distance)
}

// calculateConeGain returns the gain factor based on the angle between the
// emitter's forward direction and the direction to the listener, interpolating
// between 1.0 and outerConeGain across the inner and outer cone angles.
func calculateConeGain(
	directionZ sprec.Vec3,
	listenerOffset sprec.Vec3,
	distance float32,
	innerConeAngle sprec.Angle,
	outerConeAngle sprec.Angle,
	outerConeGain float32,
) float32 {

	if distance < 0.001 {
		return 1.0 // fully audible when very close to the emitter
	}

	dot := sprec.Vec3Dot(directionZ, sprec.Vec3Quot(listenerOffset, distance))

	relativeAngle := sprec.Acos(sprec.Clamp(dot, -1.0, 1.0))
	switch {
	case relativeAngle < innerConeAngle:
		return 1.0 // fully audible within the inner cone
	case relativeAngle > outerConeAngle:
		return outerConeGain // attenuated outside the outer cone
	default:
		ratio := float32((relativeAngle - innerConeAngle) / (outerConeAngle - innerConeAngle))
		return sprec.Mix(1.0, outerConeGain, ratio)
	}
}
