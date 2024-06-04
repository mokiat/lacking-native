/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out; // color + metallic
layout(location = 1) out vec4 fbColor1Out; // normal + roughness

/*range $line := .TextureLines */
/* $line */
/*end*/

/*if .UniformLines */
layout (std140) uniform Material
{
	/*range $line := .UniformLines */
	/* $line */
	/*end*/
};
/*end*/

/*template "ubo_camera.glsl"*/

smooth in vec3 normalInOut;
/*if .UseTexCoords*/
smooth in vec2 texCoordInOut;
/*end*/
/*if .UseVertexColoring*/
smooth in vec4 colorInOut;
/*end*/

void main()
{
	/* if not .UseVertexColoring */
	vec4 colorInOut = vec4(1.0, 1.0, 1.0, 1.0);
	/* end */

	vec4 color = vec4(1.0, 1.0, 1.0, 1.0);
	float metallic = 0.0;
	vec3 normal = normalize(normalInOut);
	float roughness = 0.0;

	/*range $line := .CodeLines */
	/* $line */
	/*end*/

	fbColor0Out = vec4(color.xyz, metallic);
	fbColor1Out = vec4(normal, roughness);
}