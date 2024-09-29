/* template "version.glsl" */

layout(location = 0) out vec4 fbColor0Out; // color + metallic
layout(location = 1) out vec4 fbColor1Out; // normal + roughness

/* range $line := .TextureLines */
/* $line */
/* end */

/* if .UniformLines */
layout (std140) uniform Material
{
	/* range $line := .UniformLines */
	/* $line */
	/* end */
};
/* end */

/* template "ubo_camera.glsl" */

smooth in vec3 normalInOut;
smooth in vec3 tangentInOut;
smooth in vec2 texCoordInOut;
smooth in vec4 colorInOut;

vec3 mapNormal(vec3 texel, float scale)
{
	vec3 ls_normal = (texel * 2.0 - vec3(1.0)) * vec3(scale, scale, 1.0);
	vec3 ws_normal = normalize(normalInOut);
	vec3 ws_tangent = normalize(tangentInOut);
	vec3 ws_bitangent = normalize(cross(ws_normal, ws_tangent));
	mat3 tbn = mat3(ws_tangent, ws_bitangent, ws_normal);
	return tbn * normalize(ls_normal);
}

void main()
{
	vec3 normal = normalize(normalInOut);
	vec4 color = colorInOut;
	float metallic = 0.0;
	float roughness = 1.0;

	/* range $line := .CodeLines */
	/* $line */
	/* end */

	fbColor0Out = vec4(color.xyz, metallic);
	fbColor1Out = vec4(normal, roughness);
}