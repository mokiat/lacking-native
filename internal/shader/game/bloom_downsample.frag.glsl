/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D lackingSourceImage;

noperspective in vec2 texCoordInOut;

// https://en.wikipedia.org/wiki/Relative_luminance
float rgbToBrightness(vec3 rgb) {
	return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}

void main()
{
	vec3 sourceHDR = texture(lackingSourceImage, texCoordInOut).xyz;
	float brightness = rgbToBrightness(sourceHDR);
	// TODO: Make ranges configurable.
	const float lower = 5.0;
	const float upper = 100.0;
	float amount = smoothstep(lower, upper, brightness) / lower;
	fbColor0Out = vec4(sourceHDR * amount, 1.0);
}
