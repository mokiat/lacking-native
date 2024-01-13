/*template "version.glsl"*/

layout(location = 0) in vec4 coordIn;
layout(location = 1) in vec3 normalIn;
/*if .UseTexturing*/
layout(location = 3) in vec2 texCoordIn;
/*end*/
/*if .UseVertexColoring*/
layout(location = 4) in vec4 colorIn;
/*end*/
/*if .UseArmature*/
layout(location = 5) in vec4 weightsIn;
layout(location = 6) in uvec4 jointsIn;
/*end*/

/*template "ubo_camera.glsl"*/

/*template "ubo_model.glsl"*/

smooth out vec3 normalInOut;
/*if .UseTexturing*/
smooth out vec2 texCoordInOut;
/*end*/
/*if .UseVertexColoring*/
smooth out vec4 colorInOut;
/*end*/

void main()
{
	/*if .UseTexturing*/
	texCoordInOut = texCoordIn;
	/*end*/
	/*if .UseVertexColoring*/
	colorInOut = colorIn;
	/*end*/
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
	vec3 worldNormal =
		inverse(transpose(mat3(modelMatrixA))) * (normalIn * weightsIn.x) +
		inverse(transpose(mat3(modelMatrixB))) * (normalIn * weightsIn.y) +
		inverse(transpose(mat3(modelMatrixC))) * (normalIn * weightsIn.z) +
		inverse(transpose(mat3(modelMatrixD))) * (normalIn * weightsIn.w);
	/*else*/
	mat4 modelMatrix = modelMatrixIn[gl_InstanceID];
	vec4 worldPosition = modelMatrix * coordIn;
	vec3 worldNormal = inverse(transpose(mat3(modelMatrix))) * normalIn;
	/*end*/
	normalInOut = worldNormal;
	gl_Position = projectionMatrixIn * (viewMatrixIn * worldPosition);
}