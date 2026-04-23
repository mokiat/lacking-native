package internal

import (
	"github.com/mokiat/gblob"
	"github.com/mokiat/lacking/audio"
)

func newOutputNode() *OutputNode {
	return &OutputNode{}
}

type OutputNode struct {
	audio.Node // marker interface

	outputData []byte
}

func (n *OutputNode) Prepare(outputData []byte) {
	n.outputData = outputData
}

func (n *OutputNode) Process(ctx ProcessContext, inputFrames, _ FrameList) {
	// NOTE: Ignoring outputFrames as they are not needed for the output node.
	// Instead, writing directly to the output buffer.
	buffer := gblob.LittleEndianBlock(n.outputData)
	for i, frame := range inputFrames {
		frame.Clamp()
		buffer.SetInt16(i*4+0, float32ToInt16(frame.Left))
		buffer.SetInt16(i*4+2, float32ToInt16(frame.Right))
	}
}
