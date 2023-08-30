/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;

// https://en.wikipedia.org/wiki/Relative_luminance
float rgbToBrightness(vec3 rgb) {
	return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}

float brightnessAt(vec3 sampleSpot, vec2 offset) {
	vec3 texRGB = texture(fbColor0TextureIn, sampleSpot.xy + offset*sampleSpot.z).xyz;
	return rgbToBrightness(texRGB);
}

#define SAMPLE_COUNT 29

void main()
{
	// The following sample spots were hand-crafted. The idea was to have them
	// somewhat randomized so that they don't lie on the same line, which causes
	// sharp exposure transitions.
	vec3 sampleSpots[SAMPLE_COUNT] = vec3[](
		vec3( 0.0,  0.0, 0.01),

		vec3(-0.015, 0.035, 0.01),
		vec3(0.03, 0.025, 0.01),
		vec3(-0.02, -0.025, 0.01),
		vec3(0.025, -0.015, 0.01),

		vec3(-0.05, 0.01, 0.02),
		vec3(-0.055, -0.06, 0.02),
		vec3(0.04, -0.06, 0.02),
		vec3(0.08, 0.05, 0.02),

		vec3(0.03, 0.1, 0.03),
		vec3(0.125, 0.07, 0.03),
		vec3(0.09, -0.03, 0.03),
		vec3(-0.01, -0.11, 0.03),
		vec3(-0.1, -0.12, 0.03),
		vec3(-0.12, -0.02, 0.03),
		vec3(-0.06, 0.08, 0.03),
		vec3(0.09, -0.125, 0.03),

		vec3(0.12, 0.2, 0.07),
		vec3(0.25, 0.03, 0.07),
		vec3(0.225, -0.18, 0.07),
		vec3(0.1, -0.34, 0.07),
		vec3(-0.075, -0.25, 0.07),
		vec3(-0.25, -0.175, 0.07),
		vec3(-0.275, 0.05, 0.07),
		vec3(-0.09, 0.2, 0.07),

		vec3(0.37, 0.265, 0.1),
		vec3(0.375, -0.21, 0.1),
		vec3(-0.31, 0.34, 0.1),
		vec3(-0.37, -0.25, 0.1)
	);

	float totalBrightness = 0.0;
	for (int i = 0; i < SAMPLE_COUNT; i++) {
		vec3 sampleSpot = sampleSpots[i] + vec3(0.5, 0.5, 0.0);

		// By taking a few samples and using their minimum value outliers are
		// reduced. This works well in situations where there is a very bright spot
		// in the frame (e.g. a lamp or the Sun).
		float brightness = brightnessAt(sampleSpot, vec2(0.0, 0.0));
		brightness = min(brightness, brightnessAt(sampleSpot, vec2(-1.0,  0.0)));
		brightness = min(brightness, brightnessAt(sampleSpot, vec2( 1.0,  0.0)));
		brightness = min(brightness, brightnessAt(sampleSpot, vec2( 0.0, -1.0)));
		brightness = min(brightness, brightnessAt(sampleSpot, vec2( 0.0,  1.0)));

		totalBrightness += brightness / float(SAMPLE_COUNT);
	}
	fbColor0Out = vec4(totalBrightness, 0.0, 0.0, 1.0);
}
