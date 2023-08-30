/*template "version.glsl"*/

layout(location = 0) out vec4 fragmentColor;

noperspective in vec4 colorInOut;

void main()
{
	fragmentColor = colorInOut;
}
