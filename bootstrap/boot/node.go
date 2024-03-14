package boot

type Node struct {
	data *nodeData
}

type nodeMap struct{}

func (nm *nodeMap) NewNode(value any) Node {
	data := &nodeData{}
	return Node{data}
}

type nodeData struct{}
