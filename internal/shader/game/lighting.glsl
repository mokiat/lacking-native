// getScreenUVCoords returns the coordinates on the screen as though there is
// a UV mapping on top (meaning {0.0, 0.0} bottom left and {1.0, 1.0} top right).
vec2 getScreenUVCoords(vec4 viewport)
{
	return (gl_FragCoord.xy - viewport.xy) / viewport.zw;
}

// getScreenNDC converts screen UV coordinates to NDC
// (Normalized Device Coordinates).
vec3 getScreenNDC(vec2 uv, sampler2D depthTexture)
{
  vec3 nonNormized = vec3(uv.x, uv.y, texture(depthTexture, uv).x);
  return nonNormized * 2.0 - vec3(1.0);
}

// getViewCoords converst the NDC coords into view coordinates.
vec3 getViewCoords(vec3 ndc, mat4 projectionMatrix)
{
  vec3 clipCoords = vec3(
		ndc.x / projectionMatrix[0][0],
		ndc.y / projectionMatrix[1][1],
		-1.0
	);
  float scale = projectionMatrix[3][2] / (projectionMatrix[2][2] + ndc.z);
  return clipCoords * scale;
}

// getWorldCoords converts the specified view coords into world coordinates
// depending on the camera positioning.
vec3 getWorldCoords(vec3 viewCoords, mat4 cameraMatrix)
{
  return (cameraMatrix * vec4(viewCoords, 1.0)).xyz;
}

// getCappedDistanceAttenuation calculates the attenuation depending on the
// distance with an upper bound on the maximum distance.
float getCappedDistanceAttenuation(float dist, float maxDist)
{
  float sqrDist = dist * dist;
  float gradient = 1.0 - dist / maxDist;
  return clamp(gradient, 0.0, 1.0) / (1.0 + sqrDist);
}

// getConeAttenuation calculates the attenuation for a cone-shaped light
// source depending on the light direction.
float getConeAttenuation(float angle, float outerAngle, float innerAngle)
{
  float hardAttenuation = 1.0 - step(outerAngle, angle);
  float softAttenuation = clamp((outerAngle - angle) / (outerAngle - innerAngle + 0.001), 0.0, 1.0);
  return hardAttenuation * (softAttenuation * softAttenuation);
}

struct FresnelInput
{
	vec3 reflectance_f0;
	vec3 half_dir;
	vec3 view_dir;
};

vec3 calculate_fresnel(FresnelInput i)
{
	float half_dot_view = clamp(dot(i.half_dir, i.view_dir), 0.0, 1.0);
	return i.reflectance_f0 + (1.0 - i.reflectance_f0) * pow(1.0 - half_dot_view, 5.0);
}

struct DistributionInput
{
	vec3 normal;
	vec3 half_dir;
	float roughness;
};

float calculate_distribution(DistributionInput i)
{
	i.roughness = clamp(i.roughness, 0.02, 1.0);
	float alpha = i.roughness * i.roughness;
	float alphaSqr = alpha * alpha;
	float halfNormDot = clamp(dot(i.normal, i.half_dir), 0.0, 1.0);
	float denom = clamp((halfNormDot * halfNormDot) * (alphaSqr - 1.0) + 1.0, 0.00001, 1.0);
	return alphaSqr / (pi * denom * denom);
}

struct geometryInput
{
	vec3 normal;
	vec3 viewDirection;
	float roughness;
};

float calculateGeometry(geometryInput i)
{
	float normViewDot = clamp(dot(i.normal, i.viewDirection), 0.01, 1.0);
	return normViewDot / (normViewDot * (1.0 - i.roughness) + i.roughness);
}

struct directionalSetup
{
	vec3 baseColor;
	float metallic;
	float roughness;
	vec3 viewDirection;
	vec3 lightDirection;
	vec3 normal;
	vec3 lightIntensity;
};

vec3 calculateDirectionalHDR(directionalSetup s)
{
	float norm_dot_view = clamp(dot(s.normal, s.viewDirection), 0.01, 1.0);
	float norm_dot_light = clamp(dot(s.normal, s.lightDirection), 0.01, 1.0);

	vec3 mid_vector = s.lightDirection + s.viewDirection;
	bool is_zero_vector = all(lessThan(abs(mid_vector), vec3(0.001)));
	vec3 half_dir = is_zero_vector ? vec3(0.0) : normalize(mid_vector);

	vec3 refracted_color = s.baseColor * (1.0 - s.metallic);
	vec3 refraction_hdr = refracted_color / pi;

	const vec3 dielectric_reflectance = vec3(0.03);
	vec3 reflected_color = mix(dielectric_reflectance, s.baseColor, s.metallic);
	vec3 fresnel = calculate_fresnel(FresnelInput(
		reflected_color,
		half_dir,
		s.viewDirection
	));
	float distribution_factor = calculate_distribution(DistributionInput(
		s.normal,
		half_dir,
		s.roughness
	));
	float geom_view_factor = calculateGeometry(geometryInput(
		s.normal,
		s.viewDirection,
		s.roughness
	));
	float geom_light_factor = calculateGeometry(geometryInput(
		s.normal,
		s.lightDirection,
		s.roughness
	));
	float geom_factor = geom_view_factor * geom_light_factor;
	float normalization_factor = 4.0 * norm_dot_view * norm_dot_light;
	vec3 reflection_hdr = vec3(distribution_factor * geom_factor / normalization_factor);

	vec3 brdf = mix(refraction_hdr, reflection_hdr, fresnel);
	return brdf * norm_dot_light * s.lightIntensity;
}

float textureClampToBorder(sampler2DArrayShadow tex, vec4 coord, float dValue)
{
	if (any(lessThan(coord.xy, vec2(0.0))) || any(greaterThan(coord.xy, vec2(1.0)))) {
		return dValue;
	}
	return texture(tex, coord);
}

struct ShadowSetup
{
	mat4 lightShadowMatrix;
	vec3 worldPosition;
	vec3 normal;
	float depth;
};

float shadowAttenuation(sampler2DArrayShadow shadowTex, ShadowSetup s)
{
	vec2 scale = vec2(1.0) / vec2(textureSize(shadowTex, 0).xy);

	float w = 64.0; // TODO: From projection matrix
	float texelSize = w * max(scale.x, scale.y);
	float bias = texelSize * 2.0;

	vec3 pointPosition = s.worldPosition + s.normal * bias;
	vec4 shadowClipPosition = s.lightShadowMatrix * vec4(pointPosition, 1.0);
	vec3 shadowNDCPosition = shadowClipPosition.xyz / shadowClipPosition.w;
	vec3 shadowUVPosition = shadowNDCPosition * 0.5 + vec3(0.5);

	vec4 texCoord = vec4(shadowUVPosition.xy, s.depth, shadowUVPosition.z);
	return textureClampToBorder(shadowTex, texCoord, 1.0);
}
