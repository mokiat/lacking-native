/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;
uniform sampler2D fbColor1TextureIn;
uniform sampler2D fbDepthTextureIn;
/*if .UseShadowMapping*/
uniform sampler2DShadow fbShadowTextureIn;
/*end*/

/*template "ubo_camera.glsl"*/

/*template "ubo_light.glsl"*/

/*template "ubo_light_properties.glsl"*/

/*template "math.glsl"*/

/*template "lighting.glsl"*/

void main()
{
	vec2 screenCoord = getScreenUVCoords(viewportIn);
	vec3 ndcPosition = getScreenNDC(screenCoord, fbDepthTextureIn);
	vec3 viewPosition = getViewCoords(ndcPosition, projectionMatrixIn);
	vec3 worldPosition = getWorldCoords(viewPosition, cameraMatrixIn);
	vec3 cameraPosition = cameraMatrixIn[3].xyz;

	vec4 albedoMetalness = texture(fbColor0TextureIn, screenCoord);
	vec4 normalRoughness = texture(fbColor1TextureIn, screenCoord);
	vec3 baseColor = albedoMetalness.xyz;
	vec3 normal = normalize(normalRoughness.xyz);
	float metalness = albedoMetalness.w;
	float roughness = normalRoughness.w;

	vec3 refractedColor = baseColor * (1.0 - metalness);
	vec3 reflectedColor = mix(vec3(0.02), baseColor, metalness);

	vec3 lightDirection = normalize(lightMatrixIn[2].xyz);

	vec3 lightIntensity = lightIntensityIn.xyz * lightIntensityIn.w;

	vec3 hdr = calculateDirectionalHDR(directionalSetup(
		roughness,
		reflectedColor,
		refractedColor,
		normalize(cameraPosition - worldPosition),
		lightDirection,
		normal,
		lightIntensity
	));

	float attenuation = 1.0;

	/*if .UseShadowMapping*/
	attenuation *= shadowAttenuation(fbShadowTextureIn, ShadowSetup(
		lightProjectionMatrixIn,
		lightViewMatrixIn,
		lightMatrixIn,
		worldPosition,
		normal
	));
	/*end*/

	fbColor0Out = vec4(hdr * attenuation, 1.0);
}
