/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor2TextureIn;

/*template "ubo_camera.glsl"*/

/*template "math.glsl"*/

/*template "lighting.glsl"*/

void main()
{
	// TODO: Emissive intensity should depend on the distance to camera,
	// according to inverse square law.
	vec2 screenCoord = getScreenUVCoords(viewportIn);
	vec4 emissiveColor = texture(fbColor2TextureIn, screenCoord);
	fbColor0Out = vec4(emissiveColor.xyz, 1.0);
}