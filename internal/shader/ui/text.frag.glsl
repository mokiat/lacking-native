/*template "version.glsl"*/

layout(location = 0) out vec4 fragmentColor;

/*template "ubo_material.glsl"*/

uniform sampler2D fontTextureIn;

noperspective in vec2 texCoordInOut;

void main()
{
	float amount = texture(fontTextureIn, texCoordInOut).x;
	fragmentColor = vec4(amount) * colorIn;
}
