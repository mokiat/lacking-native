package internal

type State struct {
	CullTest                    bool
	CullFace                    uint32
	FrontFace                   uint32
	DepthTest                   bool
	DepthMask                   bool
	DepthComparison             uint32
	StencilTest                 bool
	StencilOpStencilFailFront   uint32
	StencilOpDepthFailFront     uint32
	StencilOpPassFront          uint32
	StencilOpStencilFailBack    uint32
	StencilOpDepthFailBack      uint32
	StencilOpPassBack           uint32
	StencilComparisonFuncFront  uint32
	StencilComparisonRefFront   int32
	StencilComparisonMaskFront  uint32
	StencilComparisonFuncBack   uint32
	StencilComparisonRefBack    int32
	StencilComparisonMaskBack   uint32
	StencilMaskFront            uint32
	StencilMaskBack             uint32
	ColorMask                   [4]bool
	Blending                    bool
	BlendColor                  [4]float32
	BlendModeRGB                uint32
	BlendModeAlpha              uint32
	BlendSourceFactorRGB        uint32
	BlendDestinationFactorRGB   uint32
	BlendSourceFactorAlpha      uint32
	BlendDestinationFactorAlpha uint32
}
