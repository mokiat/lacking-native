/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;
uniform sampler2D fbColor1TextureIn;
uniform sampler2D fbDepthTextureIn;
uniform samplerCube reflectionTextureIn;
uniform samplerCube refractionTextureIn;

/*template "ubo_camera.glsl"*/

/*template "math.glsl"*/

/*template "lighting.glsl"*/

struct AmbientFresnelInput
{
	vec3 reflectance_f0;
	vec3 normal;
	vec3 viewDirection;
	float roughness;
};

vec3 calc_fresnel_ambient(AmbientFresnelInput i)
{
	float norm_dot_view = min(abs(dot(i.normal, i.viewDirection)), 1.0);
	return i.reflectance_f0 + (max(vec3(1.0 - i.roughness), i.reflectance_f0) - i.reflectance_f0) * pow(1.0 - norm_dot_view, 5.0);
}

struct ambientSetup {
	vec3 baseColor;
	float metallic;
	float roughness;
	vec3 viewDirection;
	vec3 normal;
};

vec3 calculateAmbientHDR(ambientSetup s)
{
	vec3 albedo = s.baseColor * (1.0 - s.metallic);
	vec3 irradiance_light = texture(refractionTextureIn, -s.normal).xyz;
	vec3 refraction_brdf = irradiance_light * (albedo / pi);

	vec3 reflection_f0 = mix(dielectric_reflectance, s.baseColor, s.metallic);
	vec3 fresnel = calc_fresnel_ambient(AmbientFresnelInput(
		reflection_f0,
		s.normal,
		s.viewDirection,
		s.roughness
	));
	vec3 light_dir = reflect(-s.viewDirection, s.normal);
	float max_lod = log2(float(textureSize(reflectionTextureIn, 0).x)) - 1.0;
	float alpha = s.roughness * s.roughness;
	vec3 reflected_light = textureLod(reflectionTextureIn, -light_dir, alpha * max_lod).xyz / (2.0 * pi);
	float geometry = calc_geometry(GeometryInput(
		s.normal,
		s.viewDirection,
		s.roughness
	));
	float normalization_factor = 1.0; // works best
	vec3 reflection_brdf = reflected_light * geometry / normalization_factor;

	return mix(refraction_brdf, reflection_brdf, fresnel);
}

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

	vec3 hdr = calculateAmbientHDR(ambientSetup(
		baseColor,
		metalness,
		roughness,
		normalize(cameraPosition - worldPosition),
		normal
	));
	fbColor0Out = vec4(hdr, 1.0);
}