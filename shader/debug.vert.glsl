/*template "version.glsl"*/

layout(location = 0) in vec4 coordIn;
layout(location = 4) in vec3 colorIn;

/*template "ubo_camera.glsl"*/

smooth out vec3 colorInOut;

void main()
{
  colorInOut = colorIn;
	gl_Position = projectionMatrixIn * (viewMatrixIn * coordIn);
	// move debug lines a bit forward, taking perspective into consideration
	gl_Position.z -= 0.01 * gl_Position.w;
}
