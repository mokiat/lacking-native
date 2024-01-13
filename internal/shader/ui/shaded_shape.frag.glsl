/*template "version.glsl"*/

layout(location = 0) out vec4 fragmentColor;

uniform sampler2D colorTextureIn;
uniform vec4 colorIn = vec4(1.0, 1.0, 1.0, 1.0);

noperspective in vec2 texCoordInOut;

void main()
{
	fragmentColor = texture(colorTextureIn, texCoordInOut) * colorIn;
}
