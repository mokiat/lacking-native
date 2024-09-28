/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;
uniform sampler2D fbColor1TextureIn;
uniform sampler2D fbDepthTextureIn;

/*template "ubo_camera.glsl"*/

/*template "ubo_light.glsl"*/

/*template "math.glsl"*/

/*template "lighting.glsl"*/

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

	vec3 lightDirection = lackingLightModelMatrix[3].xyz - worldPosition;
	float lightDistance = length(lightDirection);
	float lightRange = lightSpanIn.x;
	float lightOuterAngle = lightSpanIn.y;
	float lightInnerAngle = lightSpanIn.z;
	float lightAngle = acos(dot(normalize(lightDirection), normalize(lackingLightModelMatrix[1].xyz)));
	float distAttenuation = getCappedDistanceAttenuation(lightDistance, lightRange);
	float coneAttenuation = getConeAttenuation(lightAngle, lightOuterAngle, lightInnerAngle);

	vec3 lightIntensity = lightIntensityIn.xyz * lightIntensityIn.w;

	vec3 hdr = calculateDirectionalHDR(directionalSetup(
		baseColor,
		metallic,
		roughness,
		normalize(cameraPosition - worldPosition),
		normalize(lightDirection),
		normal,
		lightIntensity
	));
	fbColor0Out = vec4(hdr * distAttenuation * coneAttenuation, 1.0);
}