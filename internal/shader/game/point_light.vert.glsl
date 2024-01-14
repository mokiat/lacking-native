/*template "version.glsl"*/

layout(location = 0) in vec3 coordIn;

/*template "ubo_camera.glsl"*/

/*template "ubo_light.glsl"*/

void main()
{
	vec4 position = lightMatrixIn * vec4(coordIn, 1.0);
	gl_Position = projectionMatrixIn * (viewMatrixIn * position);
}
