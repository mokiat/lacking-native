layout (std140) uniform Light
{
	mat4 lackingLightShadowMatrices[8];
	mat4 lackingLightModelMatrix;
	vec2 lackingLightShadowCascades[8];
	vec4 lightIntensityIn;
	vec4 lightSpanIn;
};