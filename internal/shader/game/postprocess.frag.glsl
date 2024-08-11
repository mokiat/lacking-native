/*template "version.glsl"*/

layout(location = 0) out vec4 fbColor0Out;

uniform sampler2D fbColor0TextureIn;
/*if .UseBloom*/
uniform sampler2D lackingBloomTexture;
/*end*/

/*template "ubo_postprocess.glsl"*/

noperspective in vec2 texCoordInOut;

void main()
{
	float exposure = exposureIn.x;
	vec3 hdr = texture(fbColor0TextureIn, texCoordInOut).xyz;
	/*if .UseBloom*/
	hdr += texture(lackingBloomTexture, texCoordInOut).xyz;
	/*end*/
	vec3 exposedHDR = hdr * exposure;
	/*if .UseReinhard*/
	vec3 ldr = exposedHDR / (exposedHDR + vec3(1.0));
	/*end*/
	/*if .UseExponential*/
	vec3 ldr = vec3(1.0) - exp2(-exposedHDR);
	/*end*/
	fbColor0Out = vec4(ldr, 1.0);
	fbColor0Out.rgb = pow(fbColor0Out.rgb, vec3(1.0/2.2));
}
