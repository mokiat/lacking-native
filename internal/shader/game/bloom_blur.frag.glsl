/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D lackingSourceImage;

layout (std140) uniform BloomBlurData
{
	float horizontal;
};

noperspective in vec2 texCoordInOut;

#define SAMPLE_COUNT 5

void main()
{
	vec2 size = vec2(textureSize(lackingSourceImage, 0));
	vec2 hStep = vec2(1.0 / float(size.x), 0.0);
	vec2 vStep = vec2(0.0, 1.0 / float(size.y));

	vec2 sampleShiftWeights[SAMPLE_COUNT] = vec2[](
		vec2(-2.0, 1.0),
		vec2(-1.0, 2.0),
		vec2(0.0, 4.0),
		vec2(1.0, 2.0),
		vec2(2.0, 1.0)
	);

	vec3 targetHDR = vec3(0.0, 0.0, 0.0);
	for (int i = 0; i < SAMPLE_COUNT; i++) {
		vec2 sampleShiftWeight = sampleShiftWeights[i];
		vec2 offset;
		if (horizontal > 0.5) {
			offset = sampleShiftWeight.x * hStep;
		} else {
			offset = sampleShiftWeight.x * vStep;
		}
		float weight = sampleShiftWeight.y;
		targetHDR += texture(lackingSourceImage, texCoordInOut + offset).xyz * weight;
	}
	targetHDR /= 10.0;
	fbColor0Out = vec4(targetHDR, 1.0);
}
