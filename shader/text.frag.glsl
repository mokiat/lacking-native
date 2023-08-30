/*template "version.glsl"*/

layout(location = 0) out vec4 fragmentColor;

// TODO: Use binding
uniform sampler2D textureIn;
uniform vec4 colorIn = vec4(1.0, 1.0, 1.0, 1.0);

noperspective in vec2 texCoordInOut;

void main()
{
	float amount = texture(textureIn, texCoordInOut).x;
	fragmentColor = vec4(amount) * colorIn;
}
