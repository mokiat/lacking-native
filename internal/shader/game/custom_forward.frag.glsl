/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

layout (std140) uniform Material
{
	/*range $line := .UniformLines */
	/* $line */
	/*end*/
};

void main()
{
	vec4 color = vec4(0.0, 0.0, 0.0, 1.0);
	/*range $line := .CodeLines */
	/* $line */
	/*end*/
	fbColor0Out = color;
}