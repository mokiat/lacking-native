/* template "version.glsl" */

/* if .HasCoords */
layout(location = 0) in vec4 attrCoord;
/* end */
/* if .HasNormals */
layout(location = 1) in vec3 attrNormal;
/* end */
/* if .HasTangents */
layout(location = 2) in vec3 attrTangent;
/* end */
/* if .HasTexCoords */
layout(location = 3) in vec2 attrTexCoord;
/* end */
/* if .HasVertexColoring */
layout(location = 4) in vec4 attrColor;
/* end */
/* if .HasArmature */
layout(location = 5) in vec4 attrWeights;
layout(location = 6) in uvec4 attrJoints;
/* end */

/* if .LoadGeometryPreset */
/* template "preset_geometry_global.vert.glsl" */
/* end */

/* if .LoadSkyPreset */
/* template "preset_sky_global.vert.glsl" */
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
	/* if .LoadGeometryPreset */
	/* template "preset_geometry_pre.vert.glsl" */
	/* end */

	/* if .LoadSkyPreset */
	/* template "preset_sky_pre.vert.glsl" */
	/* end */

	/* range $line := .CodeLines */
	/* $line */
	/* end */
}
