/* template "version.glsl" */

layout(location = 0) in vec4 coordIn;
/* if .UseArmature */
layout(location = 5) in vec4 weightsIn;
layout(location = 6) in uvec4 jointsIn;
/* end */

/* template "ubo_camera.glsl" */

/* template "ubo_model.glsl" */

/* if .UseArmature */
/* template "ubo_armature.glsl" */
/* end */

void main()
{
	/* if .UseArmature */
	mat4 model_matrix =
		boneMatrixIn[jointsIn.x] * weightsIn.x +
		boneMatrixIn[jointsIn.y] * weightsIn.y +
		boneMatrixIn[jointsIn.z] * weightsIn.z +
		boneMatrixIn[jointsIn.w] * weightsIn.w;
	/* else */
	mat4 model_matrix =	modelMatrixIn[gl_InstanceID];
	/* end */

	// NOTE: For custom shaders: To get the model position of the vertex
	// just multiply the coordIn by the inverse model_matrix. Don't change
	// the armature to relative matrices.
	gl_Position = projectionMatrixIn * (viewMatrixIn * (model_matrix * coordIn));
}
