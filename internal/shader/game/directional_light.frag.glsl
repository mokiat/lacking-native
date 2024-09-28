/* template "version.glsl" */

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;
uniform sampler2D fbColor1TextureIn;
uniform sampler2D fbDepthTextureIn;
/* if .UseShadowMapping */
uniform sampler2DArrayShadow lackingShadowMap;
/* end */

/* template "ubo_camera.glsl" */

/* template "ubo_light.glsl" */

/* template "math.glsl" */

/* template "lighting.glsl" */

void main()
{
	vec2 screenCoord = getScreenUVCoords(viewportIn);
	vec3 ndcPosition = getScreenNDC(screenCoord, fbDepthTextureIn);
	vec3 viewPosition = getViewCoords(ndcPosition, projectionMatrixIn);
	vec3 worldPosition = getWorldCoords(viewPosition, cameraMatrixIn);
	vec3 cameraPosition = cameraMatrixIn[3].xyz;

	vec4 albedoMetallic = texture(fbColor0TextureIn, screenCoord);
	vec3 baseColor = max(albedoMetallic.xyz, vec3(0.0));
	float metallic = clamp(albedoMetallic.w, 0.0, 1.0);

	vec4 normalRoughness = texture(fbColor1TextureIn, screenCoord);
	vec3 normal = normalize(normalRoughness.xyz);
	float roughness = clamp(normalRoughness.w, 0.0, 1.0);

	vec3 lightDirection = normalize(lackingLightModelMatrix[2].xyz);
	vec3 lightIntensity = lightIntensityIn.xyz * lightIntensityIn.w;

	vec3 hdr = calculateDirectionalHDR(directionalSetup(
		baseColor,
		metallic,
		roughness,
		normalize(cameraPosition - worldPosition),
		lightDirection,
		normal,
		lightIntensity
	));

	float attenuation = 1.0;

	/* if .UseShadowMapping */
	int cascadeIndex = -1;
	for (int i = 0; i < 4; i++) {
		vec2 cascade = lackingLightShadowCascades[i];
		if ((viewPosition.z <= -cascade.x) && (viewPosition.z > -cascade.y)) {
			cascadeIndex = i;
		}
	}
	if (cascadeIndex >= 0) {
		float shadowFactor = shadowAttenuation(lackingShadowMap, ShadowSetup(
			lackingLightShadowMatrices[cascadeIndex],
			worldPosition,
			normal,
			float(cascadeIndex)
		));
		attenuation *= shadowFactor;
	}
	/* end */

	fbColor0Out = vec4(hdr * attenuation, 1.0);
}
