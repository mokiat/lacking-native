/*template "version.glsl"*/

layout(location = 0) out vec4 fragmentColor;

uniform sampler2D fontTextureIn;
uniform vec4 colorIn = vec4(1.0, 1.0, 1.0, 1.0);

noperspective in vec2 texCoordInOut;

void main()
{
	float amount = texture(fontTextureIn, texCoordInOut).x;
	fragmentColor = vec4(amount) * colorIn;
}
