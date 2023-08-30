/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform samplerCube albedoCubeTextureIn;

smooth in vec3 texCoordInOut;

void main()
{
	fbColor0Out = texture(albedoCubeTextureIn, texCoordInOut);
}
