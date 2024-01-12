/*template "version.glsl"*/

layout(location = 0) in vec4 coordIn;
/*if .UseArmature*/
layout(location = 5) in vec4 weightsIn;
layout(location = 6) in uvec4 jointsIn;
/*end*/

/*template "ubo_light.glsl"*/

/*template "ubo_model.glsl"*/

void main()
{
	/*if .UseArmature*/
	mat4 modelMatrixA = modelMatrixIn[jointsIn.x];
	mat4 modelMatrixB = modelMatrixIn[jointsIn.y];
	mat4 modelMatrixC = modelMatrixIn[jointsIn.z];
	mat4 modelMatrixD = modelMatrixIn[jointsIn.w];
	vec4 worldPosition =
		modelMatrixA * (coordIn * weightsIn.x) +
		modelMatrixB * (coordIn * weightsIn.y) +
		modelMatrixC * (coordIn * weightsIn.z) +
		modelMatrixD * (coordIn * weightsIn.w);
	/*else*/
	mat4 modelMatrix = modelMatrixIn[gl_InstanceID];
	vec4 worldPosition = modelMatrix * coordIn;
	/*end*/
  gl_Position = lightProjectionMatrixIn * (lightViewMatrixIn * worldPosition);
}
