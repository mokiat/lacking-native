/*template "version.glsl"*/

/* if .HasOutput0 */
layout(location = 0) out vec4 output0;
/* end */
/* if .HasOutput1 */
layout(location = 1) out vec4 output1;
/* end */
/* if .HasOutput2 */
layout(location = 2) out vec4 output2;
/* end */
/* if .HasOutput3 */
layout(location = 3) out vec4 output3;
/* end */

/* if .LoadGeometryPreset */
/* template "preset_geometry_global.frag.glsl" */
/* end */

/* if .LoadSkyPreset */
/* template "preset_sky_global.frag.glsl" */
/* end */


/* range $line := .TextureLines */
/* $line */
/* end */

/* if .UniformLines */
layout (std140) uniform Material
{
	/* range $line := .UniformLines */
	/* $line */
	/* end */
};
/* end */

/* range $line := .VaryingLines */
/* $line */
/* end */

void main()
{
	/* if not .HasOutput0 */
	vec4 output0 = vec4(0.0, 0.0, 0.0, 0.0);
	/* end */
	/* if not .HasOutput1 */
	vec4 output1 = vec4(0.0, 0.0, 0.0, 0.0);
	/* end */
	/* if not .HasOutput2 */
	vec4 output2 = vec4(0.0, 0.0, 0.0, 0.0);
	/* end */
	/* if not .HasOutput3 */
	vec4 output3 = vec4(0.0, 0.0, 0.0, 0.0);
	/* end */

	/* if .LoadSkyPreset */
	/* template "preset_sky_pre.frag.glsl" */
	/* end */

	/* range $line := .CodeLines */
	/* $line */
	/* end */

	/* if .LoadSkyPreset */
	/* template "preset_sky_post.frag.glsl" */
	/* end */
}
