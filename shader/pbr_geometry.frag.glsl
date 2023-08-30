/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;
layout(location = 1) out vec4 fbColor1Out;

/*if .UseAlbedoTexture*/
uniform sampler2D albedoTwoDTextureIn;
/*end*/

/*template "ubo_material.glsl"*/

smooth in vec3 normalInOut;
/*if .UseTexturing*/
smooth in vec2 texCoordInOut;
/*end*/
/*if .UseVertexColoring*/
smooth in vec4 colorInOut;
/*end*/

void main()
{
	/*if .UseAlbedoTexture*/
	vec4 color = texture(albedoTwoDTextureIn, texCoordInOut);
	/*else if .UseVertexColoring*/
	vec4 color = colorInOut;
	/*else*/
	vec4 color = albedoColorIn;
	/*end*/

  /*if .UseAlphaTest*/
	if (color.a < alphaThresholdIn) {
		discard;
	}
	/*end*/

	fbColor0Out = vec4(color.xyz, metallicIn);
	fbColor1Out = vec4(normalize(normalInOut), roughnessIn);
}
