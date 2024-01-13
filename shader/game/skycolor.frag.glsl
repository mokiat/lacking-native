/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform vec4 albedoColorIn;

void main()
{
	fbColor0Out = albedoColorIn;
}