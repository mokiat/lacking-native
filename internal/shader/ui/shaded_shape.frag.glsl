/*template "version.glsl"*/

layout(location = 0) out vec4 fragmentColor;

/*template "ubo_material.glsl"*/

uniform sampler2D colorTextureIn;

noperspective in vec2 texCoordInOut;

void main()
{
	fragmentColor = texture(colorTextureIn, texCoordInOut) * colorIn;
}
