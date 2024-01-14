/*template "version.glsl"*/

layout(location = 0) in vec2 positionIn;
layout(location = 2) in vec4 colorIn;

/*template "ubo_camera.glsl"*/

/*template "ubo_model.glsl"*/

out gl_PerVertex
{
  vec4 gl_Position;
  float gl_ClipDistance[4];
};

noperspective out vec4 colorInOut;

void main()
{
	colorInOut = colorIn;
	vec4 screenPosition = transformMatrixIn * vec4(positionIn, 0.0, 1.0);

	vec4 clipValues = clipMatrixIn * screenPosition;
	gl_ClipDistance[0] = clipValues.x;
	gl_ClipDistance[1] = clipValues.y;
	gl_ClipDistance[2] = clipValues.z;
	gl_ClipDistance[3] = clipValues.w;

	gl_Position = projectionMatrixIn * screenPosition;
}
