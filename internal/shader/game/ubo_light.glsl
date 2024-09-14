layout (std140) uniform Light
{
	mat4 lackingLightShadowMatrices[4];
	mat4 lackingLightModelMatrix;
	vec2 lackingLightShadowCascades[4];
	vec4 lightIntensityIn;
	vec4 lightSpanIn;
};