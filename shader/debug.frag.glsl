/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

smooth in vec3 colorInOut;

void main()
{
	fbColor0Out = vec4(colorInOut, 1.0);
}
