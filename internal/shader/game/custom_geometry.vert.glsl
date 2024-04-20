/*template "version.glsl"*/

layout(location = 0) in vec4 coordIn;
/* if .UseNormals */
layout(location = 1) in vec3 normalIn;
/* end */
/* if .UseTexCoords */
layout(location = 3) in vec2 texCoordIn;
/* end */
/* if .UseVertexColoring */
layout(location = 4) in vec4 colorIn;
/* end */
/* if .UseArmature */
layout(location = 5) in vec4 weightsIn;
layout(location = 6) in uvec4 jointsIn;
/* end */

/*template "ubo_camera.glsl"*/

/*template "ubo_model.glsl"*/

/*if .UseArmature*/
/*template "ubo_armature.glsl"*/
/*end*/

smooth out vec3 normalInOut;
/*if .UseTexCoords*/
smooth out vec2 texCoordInOut;
/*end*/
/*if .UseVertexColoring*/
smooth out vec4 colorInOut;
/*end*/

void main()
{
	/*if .UseNormals*/
	vec3 normal = normalIn;
	/*else*/
	vec3 normal = vec3(0.0, 1.0, 0.0);
	/*end*/

	/*if .UseTexCoords*/
	texCoordInOut = texCoordIn;
	/*end*/
	/*if .UseVertexColoring*/
	colorInOut = colorIn;
	/*end*/
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
	vec3 worldNormal =
		inverse(transpose(mat3(boneMatrixA))) * (normal * weightsIn.x) +
		inverse(transpose(mat3(boneMatrixB))) * (normal * weightsIn.y) +
		inverse(transpose(mat3(boneMatrixC))) * (normal * weightsIn.z) +
		inverse(transpose(mat3(boneMatrixD))) * (normal * weightsIn.w);
	/*else*/
	mat4 modelMatrix = modelMatrixIn[gl_InstanceID];
	vec4 worldPosition = modelMatrix * coordIn;
	vec3 worldNormal = inverse(transpose(mat3(modelMatrix))) * normal;
	/*end*/
	// NOTE: For custom shaders: To get the model position of the vertex
	// just multiply the coordIn by the inverse modelMatrixIn. Don't change
	// the armature to relative matrices.
	normalInOut = worldNormal;
	gl_Position = projectionMatrixIn * (viewMatrixIn * worldPosition);
}
