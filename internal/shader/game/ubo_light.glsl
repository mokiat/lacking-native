layout (std140) uniform Light
{
	mat4 lackingLightShadowMatrixNear;
	mat4 lackingLightShadowMatrixMid;
	mat4 lackingLightShadowMatrixFar;
	mat4 lackingLightModelMatrix;
	vec4 lackingLightShadowCascades;
	vec4 lightIntensityIn;
	vec4 lightSpanIn;
};