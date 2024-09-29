/* template "version.glsl" */

layout(location = 0) in vec4 coordIn;
/* if .UseNormals */
layout(location = 1) in vec3 normalIn;
/* end */
/* if .UseTangents */
layout(location = 2) in vec3 tangentIn;
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

/* template "ubo_camera.glsl" */

/* template "ubo_model.glsl" */

/* if .UseArmature */
/* template "ubo_armature.glsl" */
/* end */

smooth out vec3 normalInOut;
smooth out vec3 tangentInOut;
smooth out vec2 texCoordInOut;
smooth out vec4 colorInOut;

void main()
{
	/* if .UseNormals */
	vec3 ls_normal = normalIn;
	/* else */
	vec3 ls_normal = vec3(0.0, 0.0, 1.0);
	/* end */
	/* if .UseTangents */
	vec3 ls_tangent = tangentIn;
	/* else */
	vec3 ls_tangent = vec3(1.0, 0.0, 0.0);
	/* end */
	/* if .UseTexCoords */
	vec2 tex_coord = texCoordIn;
	/* else */
	vec2 tex_coord = vec2(0.0, 0.0);
	/* end */
	/* if .UseVertexColoring */
	vec4 color = colorIn;
	/* else */
	vec4 color = vec4(1.0);
	/* end */

	/* if .UseArmature */
	mat4 model_matrix =
		boneMatrixIn[jointsIn.x] * weightsIn.x + 
		boneMatrixIn[jointsIn.y] * weightsIn.y +
		boneMatrixIn[jointsIn.z] * weightsIn.z +
		boneMatrixIn[jointsIn.w] * weightsIn.w;
	/* else */
	mat4 model_matrix =	modelMatrixIn[gl_InstanceID];
	/* end */
	mat3 model_rot_matrix = inverse(transpose(mat3(model_matrix)));

	// NOTE: For custom shaders: To get the model position of the vertex
	// just multiply the coordIn by the inverse model_matrix. Don't change
	// the armature to relative matrices.
	normalInOut = model_rot_matrix * ls_normal;
	tangentInOut = model_rot_matrix * ls_tangent;
	texCoordInOut = tex_coord;
	colorInOut = color;
	gl_Position = projectionMatrixIn * (viewMatrixIn * (model_matrix * coordIn));
}
