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

		constructor:  newSnapshotConstructor(),
		tempSnapshot: new(ProcessingSnapshot),

		pendingSnapshot: new(ProcessingSnapshot),
		activeSnapshot:  new(ProcessingSnapshot),
		swapSnapshots:   false,
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

	constructor  *snapshotConstructor
	tempSnapshot *ProcessingSnapshot

	snapshotMU      sync.Mutex
	pendingSnapshot *ProcessingSnapshot
	activeSnapshot  *ProcessingSnapshot
	swapSnapshots   bool
}

func (g *Graph) Register(element Node, isOutput bool) {
	if _, exists := g.nodeMapping[element]; exists {
		panic("attempting to register node twice")
	}
	node, nodeIndex := g.allocateNode()
	node.processor = element
	node.nextNode = -1
	node.firstInboundLink = -1
	node.firstOutboundLink = -1
	node.isOutput = isOutput
	node.isWired = isOutput
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
	g.invalidateNode(srcNodeIndex)
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
	g.invalidateNode(srcNodeIndex)
	g.invalidateSnapshot()
}

func (g *Graph) Snapshot() *ProcessingSnapshot {
	g.snapshotMU.Lock()
	defer g.snapshotMU.Unlock()
	if g.swapSnapshots {
		g.swapSnapshots = false
		g.activeSnapshot, g.pendingSnapshot = g.pendingSnapshot, g.activeSnapshot
	}
	return g.activeSnapshot
}

func (g *Graph) allocateNode() (*graphNode, int32) {
	var nodeIndex int32
	if g.freeNodeIndex != -1 {
		nodeIndex = g.freeNodeIndex
		node := &g.nodes[g.freeNodeIndex]
		g.freeNodeIndex = node.nextNode // this needs to happen here
		node.processor = nil
		node.nextNode = -1
		node.firstInboundLink = -1
		node.firstOutboundLink = -1
	} else {
		nodeIndex = int32(len(g.nodes))
		g.nodes = append(g.nodes, graphNode{
			processor:         nil,
			nextNode:          -1,
			firstInboundLink:  -1,
			firstOutboundLink: -1,
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
	node.isOutput = false
	node.isWired = false
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
		g.invalidateNode(srcNodeIndex)
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
	node.firstOutboundLink = -1
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

func (g *Graph) invalidateNode(nodeIndex int32) {
	node := &g.nodes[nodeIndex]
	oldIsWired := node.isWired

	newIsWired := node.isOutput
	outLinkIndex := node.firstOutboundLink
	for !newIsWired && outLinkIndex != -1 {
		outLink := &g.outboundLinks[outLinkIndex]
		targetNodeIndex := outLink.targetNode
		targetNode := &g.nodes[targetNodeIndex]
		newIsWired = newIsWired || (targetNode.isWired || targetNode.isOutput)
		outLinkIndex = outLink.nextLink
	}
	node.isWired = newIsWired

	if oldIsWired != newIsWired {
		inLinkIndex := node.firstInboundLink
		for inLinkIndex != -1 {
			inLink := &g.inboundLinks[inLinkIndex]
			sourceNodeIndex := inLink.sourceNode
			g.invalidateNode(sourceNodeIndex)
			inLinkIndex = inLink.nextLink
		}
	}
}

func (g *Graph) invalidateSnapshot() {
	g.constructor.Construct(g, g.tempSnapshot)

	g.snapshotMU.Lock()
	defer g.snapshotMU.Unlock()
	g.swapSnapshots = true
	g.tempSnapshot, g.pendingSnapshot = g.pendingSnapshot, g.tempSnapshot
}

type graphNode struct {
	processor         Processor
	nextNode          int32
	firstInboundLink  int32
	firstOutboundLink int32
	isOutput          bool
	isWired           bool
}

type graphInboundLink struct {
	sourceNode int32
	nextLink   int32
}

type graphOutboundLink struct {
	targetNode int32
	nextLink   int32
}

func newSnapshotConstructor() *snapshotConstructor {
	return &snapshotConstructor{
		leafNodes:     ds.NewStack[int32](128),
		inputCounts:   make(map[int32]int, 128),
		nodePlacement: make(map[int32]uint32, 128),
		orderedNodes:  make([]int32, 0, 128),
	}
}

type snapshotConstructor struct {
	leafNodes     *ds.Stack[int32]
	inputCounts   map[int32]int
	nodePlacement map[int32]uint32
	orderedNodes  []int32
}

func (c *snapshotConstructor) Construct(graph *Graph, target *ProcessingSnapshot) {
	c.leafNodes.Clear()
	clear(c.inputCounts)
	clear(c.nodePlacement)
	c.orderedNodes = c.orderedNodes[:0]

	// Initialize pending nodes.
	for i := range int32(len(graph.nodes)) {
		node := &graph.nodes[i]
		if node.processor == nil {
			continue // deleted node
		}
		if !node.isWired && !node.isOutput {
			continue // not connected to any output
		}
		inputCount := graph.countInboundLinks(i)
		if inputCount > 0 {
			c.inputCounts[i] = inputCount
		} else {
			c.leafNodes.Push(i)
		}
	}

	// Run Kahn's algorithm to find a topological ordering of the nodes.
	for !c.leafNodes.IsEmpty() {

		// Fetch a leaf node.
		leafNodeIndex := c.leafNodes.Pop()

		// Record the node's placement in the processing order.
		c.nodePlacement[leafNodeIndex] = uint32(len(c.orderedNodes))
		c.orderedNodes = append(c.orderedNodes, leafNodeIndex)

		// Decrease the input count of the node's outbound neighbors and add new
		// leaf nodes to the stack.
		leafNode := &graph.nodes[leafNodeIndex]
		outLinkIndex := leafNode.firstOutboundLink
		for outLinkIndex != -1 {
			outLink := &graph.outboundLinks[outLinkIndex]
			targetNodeIndex := outLink.targetNode
			if inputCount, ok := c.inputCounts[targetNodeIndex]; ok {
				if inputCount > 1 {
					c.inputCounts[targetNodeIndex] = inputCount - 1
				} else {
					c.leafNodes.Push(targetNodeIndex)
					delete(c.inputCounts, targetNodeIndex)
				}
			}
			outLinkIndex = outLink.nextLink
		}
	}

	// Check for cycles.
	if len(c.inputCounts) > 0 {
		panic("audio graph has cycles")
	}

	// Reset target snapshot.
	target.Processings = target.Processings[:0]
	target.Assignments = target.Assignments[:0]

	// Create processing units for all applicable nodes.
	for _, nodeIndex := range c.orderedNodes {
		node := &graph.nodes[nodeIndex]
		target.Processings = append(target.Processings, ProcessingUnit{
			Processor:        node.processor,
			InputCount:       0,
			OutputCount:      0,
			AssignmentOffset: 0,
		})
	}

	// Create the applicable assignments.
	for sourceUnitIndex, nodeIndex := range c.orderedNodes {
		sourceUnit := &target.Processings[sourceUnitIndex]
		sourceUnit.AssignmentOffset = uint32(len(target.Assignments))

		node := &graph.nodes[nodeIndex]
		outLinkIndex := node.firstOutboundLink
		for outLinkIndex != -1 {
			outLink := &graph.outboundLinks[outLinkIndex]
			targetNodeIndex := outLink.targetNode
			if targetUnitIndex, ok := c.nodePlacement[targetNodeIndex]; ok {
				sourceUnit.OutputCount++
				targetUnit := &target.Processings[targetUnitIndex]
				targetUnit.InputCount++
				target.Assignments = append(target.Assignments, targetUnitIndex)
			}
			outLinkIndex = outLink.nextLink
		}
	}
}
