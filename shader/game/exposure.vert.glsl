/*template "version.glsl"*/

layout(location = 0) in vec2 coordIn;

void main()
{
	gl_Position = vec4(coordIn, 0.0, 1.0);
}
