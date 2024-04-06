/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

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

smooth in vec3 texCoordInOut;

void main()
{
	/*range $line := .CodeLines */
	/* $line */
	/*end*/
}