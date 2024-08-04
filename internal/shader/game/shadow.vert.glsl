/*template "version.glsl"*/

layout(location = 0) in vec4 coordIn;
/*if .UseArmature*/
layout(location = 5) in vec4 weightsIn;
layout(location = 6) in uvec4 jointsIn;
/*end*/

/*template "ubo_camera.glsl"*/

/*template "ubo_model.glsl"*/

/*if .UseArmature*/
/*template "ubo_armature.glsl"*/
/*end*/

void main()
{
	/*if .UseArmature*/
	mat4 boneMatrixA = boneMatrixIn[jointsIn.x];
	mat4 boneMatrixB = boneMatrixIn[jointsIn.y];
	mat4 boneMatrixC = boneMatrixIn[jointsIn.z];
	mat4 boneMatrixD = boneMatrixIn[jointsIn.w];
	vec4 worldPosition =
		boneMatrixA * (coordIn * weightsIn.x) +
		boneMatrixB * (coordIn * weightsIn.y) +
		boneMatrixC * (coordIn * weightsIn.z) +
		boneMatrixD * (coordIn * weightsIn.w);
	/*else*/
	mat4 modelMatrix = modelMatrixIn[gl_InstanceID];
	vec4 worldPosition = modelMatrix * coordIn;
	/*end*/
	// NOTE: For custom shaders: To get the model position of the vertex
	// just multiply the coordIn by the inverse modelMatrixIn. Don't change
	// the armature to relative matrices.
	gl_Position = projectionMatrixIn * (viewMatrixIn * worldPosition);
}
