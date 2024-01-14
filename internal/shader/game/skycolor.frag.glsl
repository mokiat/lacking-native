/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

/*template "ubo_skybox.glsl"*/

void main()
{
	fbColor0Out = albedoColorIn;
}
