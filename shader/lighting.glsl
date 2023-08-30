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

struct fresnelInput
{
	vec3 reflectanceF0;
	vec3 halfDirection;
	vec3 lightDirection;
};

vec3 calculateFresnel(fresnelInput i)
{
	float halfLightDot = clamp(abs(dot(i.halfDirection, i.lightDirection)), 0.0, 1.0);
	return i.reflectanceF0 + (1.0 - i.reflectanceF0) * pow(1.0 - halfLightDot, 5.0);
}

struct distributionInput
{
	float roughness;
	vec3 normal;
	vec3 halfDirection;
};

float calculateDistribution(distributionInput i)
{
	float sqrRough = i.roughness * i.roughness;
	float halfNormDot = dot(i.normal, i.halfDirection);
	float denom = halfNormDot * halfNormDot * (sqrRough - 1.0) + 1.0;
	return sqrRough / (pi * denom * denom);
}

struct geometryInput
{
	float roughness;
};

float calculateGeometry(geometryInput i)
{
	// TODO: Use better model
	return 1.0 / 4.0;
}

struct directionalSetup
{
	float roughness;
	vec3 reflectedColor;
	vec3 refractedColor;
	vec3 viewDirection;
	vec3 lightDirection;
	vec3 normal;
	vec3 lightIntensity;
};

vec3 calculateDirectionalHDR(directionalSetup s)
{
	vec3 halfDirection = normalize(s.lightDirection + s.viewDirection);
	vec3 fresnel = calculateFresnel(fresnelInput(
		s.reflectedColor,
		halfDirection,
		s.lightDirection
	));
	float distributionFactor = calculateDistribution(distributionInput(
		s.roughness,
		s.normal,
		halfDirection
	));
	float geometryFactor = calculateGeometry(geometryInput(
		s.roughness
	));
	vec3 reflectedHDR = fresnel * distributionFactor * geometryFactor;
	vec3 refractedHDR = (vec3(1.0) - fresnel) * s.refractedColor / pi;
	return (reflectedHDR + refractedHDR) * s.lightIntensity * clamp(dot(s.normal, s.lightDirection), 0.0, 1.0);
}

struct ShadowSetup
{
	mat4 lightProjectionMatrix;
	mat4 lightViewMatrix;
	mat4 lightMatrix;
	vec3 worldPosition;
	vec3 normal;
};

float shadowAttenuation(sampler2DShadow shadowTex, ShadowSetup s)
{
	vec2 scale = vec2(1.0) / vec2(textureSize(shadowTex, 0));

	float w = 64.0; // TODO: From projection matrix
	float texelSize = w * max(scale.x, scale.y);
	float bias = texelSize * 2.0;

	vec3 pointPosition = s.worldPosition + s.normal * bias;
	vec4 shadowClipPosition = s.lightProjectionMatrix * (s.lightViewMatrix * vec4(pointPosition, 1.0));
	vec3 shadowNDCPosition = shadowClipPosition.xyz / shadowClipPosition.w;
	vec3 shadowUVPosition = shadowNDCPosition * 0.5 + vec3(0.5);

	const vec2[9] shifts = vec2[](
		vec2(0.00, 0.00),
		vec2(1.00, 0.00),
		vec2(0.71, 0.71),
		vec2(0.00, 1.00),
		vec2(-0.71, 0.71),
		vec2(-1.00, 0.00),
		vec2(-0.71, -0.71),
		vec2(-0.00, -1.00),
		vec2(0.71, -0.71)
	);

	float smoothness = 0.0;
	for (int i = 0; i < 9; i++) {
		vec3 offset = vec3(shifts[i] / vec2(4096), 0.0);
		smoothness += texture(shadowTex, shadowUVPosition.xyz + offset);
	}
	smoothness /= 9.0;

	float amount = texture(shadowTex, shadowUVPosition.xyz);
	amount *= smoothness;
	return amount;
}