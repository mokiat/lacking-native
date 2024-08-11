// We optimize by using vertex coords as cube texture coords.
// Additionally, we need to flip the coords. OpenGL uses renderman coordinate
// system for cube maps, contrary to the rest of the OpenGL API.
varyingDirection = -attrCoord.xyz;

// ensure that translations are ignored by setting w to 0.0
vec4 viewPosition = viewMatrixIn * vec4(attrCoord.xyz, 0.0);

// restore w to 1.0 so that projection works
vec4 position = projectionMatrixIn * vec4(viewPosition.xyz, 1.0);

// set z to w so that it has maximum depth (1.0) after projection division
gl_Position = vec4(position.xy, position.w, position.w);