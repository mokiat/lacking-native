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
	// TODO: Configurable range.
	vec3 targetHDR = smoothstep(5.0, 10.0, brightness) * sourceHDR;
	fbColor0Out = vec4(targetHDR, 1.0);
}
