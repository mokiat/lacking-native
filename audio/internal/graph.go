package internal

import (
	"sync"

	"github.com/mokiat/gog/ds"
)

func NewGraph() *Graph {
	return &Graph{
		nodeMapping: make(map[Node]int32),

		freeNodeIndex:         -1,
		freeInboundLinkIndex:  -1,
		freeOutboundLinkIndex: -1,

		tempSnapshot:    new(ProcessingSnapshot),
		pendingSnapshot: new(ProcessingSnapshot),
		activeSnapshot:  new(ProcessingSnapshot),
		swapSnapshot:    false,
	}
}

type Graph struct {
	nodes         []graphNode
	inboundLinks  []graphInboundLink
	outboundLinks []graphOutboundLink

	nodeMapping map[Node]int32

	freeNodeIndex         int32
	freeInboundLinkIndex  int32
	freeOutboundLinkIndex int32

	snapshotMU      sync.Mutex
	tempSnapshot    *ProcessingSnapshot
	pendingSnapshot *ProcessingSnapshot
	activeSnapshot  *ProcessingSnapshot
	swapSnapshot    bool
}

func (g *Graph) Register(element Node) {
	if _, exists := g.nodeMapping[element]; exists {
		panic("attempting to register node twice")
	}
	node, nodeIndex := g.allocateNode()
	node.processor = element
	node.nextNode = -1
	node.firstInboundLink = -1
	node.firstOutboundLink = -1
	g.nodeMapping[element] = nodeIndex
}

func (g *Graph) Unregister(element Node) {
	nodeIndex, ok := g.nodeMapping[element]
	if !ok {
		panic("attempting to unregister unknown node")
	}
	g.removeNodeLinks(nodeIndex)
	g.freeNode(nodeIndex)
	delete(g.nodeMapping, element)
	g.invalidateSnapshot()
}

func (g *Graph) Connect(source, target Node) {
	srcNodeIndex, ok := g.nodeMapping[source]
	if !ok {
		panic("attempting to connect from unknown source node")
	}
	tgtNodeIndex, ok := g.nodeMapping[target]
	if !ok {
		panic("attempting to connect to unknown target node")
	}
	if g.hasConnection(srcNodeIndex, tgtNodeIndex) {
		logger.Warn("Attempting to create duplicate connection in audio graph.")
		return
	}
	g.addOutboundLink(srcNodeIndex, tgtNodeIndex)
	g.addInboundLink(tgtNodeIndex, srcNodeIndex)
	g.invalidateSnapshot()
}

func (g *Graph) Disconnect(source, target Node) {
	srcNodeIndex, ok := g.nodeMapping[source]
	if !ok {
		panic("attempting to disconnect from unknown source node")
	}
	tgtNodeIndex, ok := g.nodeMapping[target]
	if !ok {
		panic("attempting to disconnect to unknown target node")
	}
	if !g.hasConnection(srcNodeIndex, tgtNodeIndex) {
		logger.Warn("Attempting to remove non-existing connection in audio graph.")
		return
	}
	g.removeOutboundLink(srcNodeIndex, tgtNodeIndex)
	g.removeInboundLink(tgtNodeIndex, srcNodeIndex)
	g.invalidateSnapshot()
}

func (g *Graph) Snapshot() *ProcessingSnapshot {
	g.snapshotMU.Lock()
	defer g.snapshotMU.Unlock()
	if g.swapSnapshot {
		g.swapSnapshot = false
		g.activeSnapshot, g.pendingSnapshot = g.pendingSnapshot, g.activeSnapshot
	}
	return g.activeSnapshot
}

func (g *Graph) allocateNode() (*graphNode, int32) {
	var nodeIndex int32
	if g.freeNodeIndex != -1 {
		nodeIndex = g.freeNodeIndex
		node := &g.nodes[g.freeNodeIndex]
		g.freeNodeIndex = node.nextNode
		node.nextNode = -1
	} else {
		nodeIndex = int32(len(g.nodes))
		g.nodes = append(g.nodes, graphNode{
			nextNode: -1,
		})
	}
	return &g.nodes[nodeIndex], nodeIndex
}

func (g *Graph) freeNode(nodeIndex int32) {
	node := &g.nodes[nodeIndex]
	node.processor = nil
	node.firstInboundLink = -1
	node.firstOutboundLink = -1
	node.nextNode = g.freeNodeIndex
	g.freeNodeIndex = nodeIndex
}

func (g *Graph) allocateInboundLink() (*graphInboundLink, int32) {
	var linkIndex int32
	if g.freeInboundLinkIndex != -1 {
		linkIndex = g.freeInboundLinkIndex
		link := &g.inboundLinks[g.freeInboundLinkIndex]
		g.freeInboundLinkIndex = link.nextLink
		link.nextLink = -1
	} else {
		linkIndex = int32(len(g.inboundLinks))
		g.inboundLinks = append(g.inboundLinks, graphInboundLink{
			nextLink: -1,
		})
	}
	return &g.inboundLinks[linkIndex], linkIndex
}

func (g *Graph) freeInboundLink(linkIndex int32) {
	link := &g.inboundLinks[linkIndex]
	link.sourceNode = -1
	link.nextLink = g.freeInboundLinkIndex
	g.freeInboundLinkIndex = linkIndex
}

func (g *Graph) addInboundLink(tgtNodeIndex, srcNodeIndex int32) {
	tgtNode := &g.nodes[tgtNodeIndex]
	inLink, inLinkIndex := g.allocateInboundLink()
	inLink.nextLink = tgtNode.firstInboundLink
	tgtNode.firstInboundLink = inLinkIndex
	inLink.sourceNode = srcNodeIndex
}

func (g *Graph) removeInboundLink(tgtNodeIndex, srcNodeIndex int32) {
	targetNode := &g.nodes[tgtNodeIndex]
	var prevLink *graphInboundLink
	currentLinkIndex := targetNode.firstInboundLink
	for currentLinkIndex != -1 {
		currentLink := &g.inboundLinks[currentLinkIndex]
		if currentLink.sourceNode == srcNodeIndex {
			if prevLink == nil {
				targetNode.firstInboundLink = currentLink.nextLink
			} else {
				prevLink.nextLink = currentLink.nextLink
			}
			g.freeInboundLink(currentLinkIndex)
			return
		}
		prevLink = currentLink
		currentLinkIndex = currentLink.nextLink
	}
}

func (g *Graph) countInboundLinks(nodeIndex int32) int {
	node := &g.nodes[nodeIndex]
	var count int
	currentLinkIndex := node.firstInboundLink
	for currentLinkIndex != -1 {
		count++
		currentLink := &g.inboundLinks[currentLinkIndex]
		currentLinkIndex = currentLink.nextLink
	}
	return count
}

func (g *Graph) allocateOutboundLink() (*graphOutboundLink, int32) {
	var linkIndex int32
	if g.freeOutboundLinkIndex != -1 {
		linkIndex = g.freeOutboundLinkIndex
		link := &g.outboundLinks[g.freeOutboundLinkIndex]
		g.freeOutboundLinkIndex = link.nextLink
		link.nextLink = -1
	} else {
		linkIndex = int32(len(g.outboundLinks))
		g.outboundLinks = append(g.outboundLinks, graphOutboundLink{
			nextLink: -1,
		})
	}
	return &g.outboundLinks[linkIndex], linkIndex
}

func (g *Graph) freeOutboundLink(linkIndex int32) {
	link := &g.outboundLinks[linkIndex]
	link.targetNode = -1
	link.nextLink = g.freeOutboundLinkIndex
	g.freeOutboundLinkIndex = linkIndex
}

func (g *Graph) addOutboundLink(srcNodeIndex, tgtNodeIndex int32) {
	srcNode := &g.nodes[srcNodeIndex]
	outLink, outLinkIndex := g.allocateOutboundLink()
	outLink.nextLink = srcNode.firstOutboundLink
	srcNode.firstOutboundLink = outLinkIndex
	outLink.targetNode = tgtNodeIndex
}

func (g *Graph) removeOutboundLink(srcNodeIndex, tgtNodeIndex int32) {
	sourceNode := &g.nodes[srcNodeIndex]
	var prevLink *graphOutboundLink
	currentLinkIndex := sourceNode.firstOutboundLink
	for currentLinkIndex != -1 {
		currentLink := &g.outboundLinks[currentLinkIndex]
		if currentLink.targetNode == tgtNodeIndex {
			if prevLink == nil {
				sourceNode.firstOutboundLink = currentLink.nextLink
			} else {
				prevLink.nextLink = currentLink.nextLink
			}
			g.freeOutboundLink(currentLinkIndex)
			return
		}
		prevLink = currentLink
		currentLinkIndex = currentLink.nextLink
	}
}

func (g *Graph) removeNodeLinks(nodeIndex int32) {
	node := &g.nodes[nodeIndex]

	// Remove reciprocal outbound links.
	inLinkIndex := node.firstInboundLink
	for inLinkIndex != -1 {
		inLink := &g.inboundLinks[inLinkIndex]
		srcNodeIndex := inLink.sourceNode
		g.removeOutboundLink(srcNodeIndex, nodeIndex)
		inLinkIndex = inLink.nextLink
	}

	// Remove reciprocal inbound links.
	outLinkIndex := node.firstOutboundLink
	for outLinkIndex != -1 {
		outLink := &g.outboundLinks[outLinkIndex]
		tgtNodeIndex := outLink.targetNode
		g.removeInboundLink(tgtNodeIndex, nodeIndex)
		outLinkIndex = outLink.nextLink
	}

	// Free inbound links.
	inLinkIndex = node.firstInboundLink
	for inLinkIndex != -1 {
		inLink := &g.inboundLinks[inLinkIndex]
		nextInLinkIndex := inLink.nextLink
		g.freeInboundLink(inLinkIndex)
		inLinkIndex = nextInLinkIndex
	}
	node.firstInboundLink = -1

	// Free outbound links.
	outLinkIndex = node.firstOutboundLink
	for outLinkIndex != -1 {
		outLink := &g.outboundLinks[outLinkIndex]
		nextOutLinkIndex := outLink.nextLink
		g.freeOutboundLink(outLinkIndex)
		outLinkIndex = nextOutLinkIndex
	}
}

func (g *Graph) hasConnection(srcNodeIndex, tgtNodeIndex int32) bool {
	srcNode := &g.nodes[srcNodeIndex]
	outLinkIndex := srcNode.firstOutboundLink
	for outLinkIndex != -1 {
		outLink := &g.outboundLinks[outLinkIndex]
		if outLink.targetNode == tgtNodeIndex {
			return true
		}
		outLinkIndex = outLink.nextLink
	}
	return false
}

func (g *Graph) invalidateSnapshot() {
	// TODO: Cache this!
	inputCounts := make(map[int32]int)
	for i := range int32(len(g.nodes)) {
		if node := &g.nodes[i]; node.processor != nil {
			inputCounts[i] = g.countInboundLinks(i)
		}
	}

	// TODO: Cache this!
	leafNodes := ds.NewStack[int32](len(g.nodes))
	for node, count := range inputCounts {
		if count == 0 {
			leafNodes.Push(node)
			delete(inputCounts, node)
		}
	}

	// TODO: Cache this!
	placementIndices := make(map[int32]uint32)

	// TODO: Cache this!
	nodeIndices := make([]int32, 0, len(g.nodes))

	snapshot := g.tempSnapshot
	snapshot.Processings = snapshot.Processings[:0]
	snapshot.Assignments = snapshot.Assignments[:0]

	for !leafNodes.IsEmpty() {
		leafNodeIndex := leafNodes.Pop()
		nodeIndices = append(nodeIndices, leafNodeIndex)
		leafNode := &g.nodes[leafNodeIndex]

		placementIndices[leafNodeIndex] = uint32(len(snapshot.Processings))
		snapshot.Processings = append(snapshot.Processings, ProcessingUnit{
			Processor:   leafNode.processor,
			InputCount:  0,
			OutputCount: 0,
		})

		outLinkIndex := leafNode.firstOutboundLink
		for outLinkIndex != -1 {
			outLink := &g.outboundLinks[outLinkIndex]
			targetNodeIndex := outLink.targetNode
			inputCounts[targetNodeIndex]--
			if inputCounts[targetNodeIndex] == 0 {
				leafNodes.Push(targetNodeIndex)
				delete(inputCounts, targetNodeIndex)
			}
			outLinkIndex = outLink.nextLink
		}
	}

	if len(inputCounts) > 0 {
		panic("audio graph has cycles")
	}

	for i := range snapshot.Processings {
		sourceUnit := &snapshot.Processings[i]
		sourceUnit.AssignmentOffset = uint32(len(snapshot.Assignments))
		nodeIndex := nodeIndices[i]
		node := &g.nodes[nodeIndex]

		outLinkIndex := node.firstOutboundLink
		for outLinkIndex != -1 {
			outLink := &g.outboundLinks[outLinkIndex]
			targetNodeIndex := outLink.targetNode
			if targetPlacementIndex, ok := placementIndices[targetNodeIndex]; ok {
				sourceUnit.OutputCount++
				targetUnit := &snapshot.Processings[targetPlacementIndex]
				targetUnit.InputCount++
				snapshot.Assignments = append(snapshot.Assignments, targetPlacementIndex)
			}
			outLinkIndex = outLink.nextLink
		}
	}

	g.snapshotMU.Lock()
	defer g.snapshotMU.Unlock()
	g.swapSnapshot = true
	g.tempSnapshot, g.pendingSnapshot = g.pendingSnapshot, g.tempSnapshot
}

type graphNode struct {
	processor         Processor
	nextNode          int32
	firstInboundLink  int32
	firstOutboundLink int32
}

type graphInboundLink struct {
	sourceNode int32
	nextLink   int32
}

type graphOutboundLink struct {
	targetNode int32
	nextLink   int32
}
