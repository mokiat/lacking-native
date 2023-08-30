/*template "version.glsl"*/

layout(location = 0) in vec3 coordIn;

/*template "ubo_camera.glsl"*/

/*template "ubo_light.glsl"*/

void main()
{
	// Due to cone shape offset.
	vec4 adjustment = vec4(0.0, -1.0, 0.0, 0.0);
	vec4 position = lightMatrixIn * (vec4(coordIn, 1.0) + adjustment);
	gl_Position = projectionMatrixIn * (viewMatrixIn * position);
}
