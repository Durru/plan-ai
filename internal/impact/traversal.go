package impact

func (g *Graph) lookupNode(id string) (Node, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

// Traverse performs BFS from a starting node up to the given depth.
// A depth of 0 returns only the starting node. A depth of -1 means unlimited.
func (g *Graph) Traverse(nodeID string, depth int) []Node {
	if _, ok := g.lookupNode(nodeID); !ok {
		return nil
	}

	adj := g.buildAdjacency()

	visited := make(map[string]bool)
	var result []Node
	queue := []struct {
		id    string
		depth int
	}{{id: nodeID, depth: 0}}
	visited[nodeID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if n, ok := g.lookupNode(current.id); ok {
			result = append(result, n)
		}
		if depth >= 0 && current.depth >= depth {
			continue
		}

		for _, neighborID := range adj[current.id] {
			if !visited[neighborID] {
				visited[neighborID] = true
				queue = append(queue, struct {
					id    string
					depth int
				}{id: neighborID, depth: current.depth + 1})
			}
		}
	}
	return result
}

// AffectedEntities traverses from a change/decision root and returns
// all entities affected (transitively reachable via edges).
func (g *Graph) AffectedEntities(rootID string) []Node {
	if _, ok := g.lookupNode(rootID); !ok {
		return nil
	}

	adj := g.buildAdjacency()

	visited := make(map[string]bool)
	var result []Node
	queue := []string{rootID}
	visited[rootID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if n, ok := g.lookupNode(current); ok {
			if current != rootID {
				result = append(result, n)
			}
		}

		for _, neighborID := range adj[current] {
			if !visited[neighborID] {
				visited[neighborID] = true
				queue = append(queue, neighborID)
			}
		}
	}
	return result
}

// IsTransitiveAffected checks whether targetID is reachable
// from sourceID via any edge path.
func (g *Graph) IsTransitiveAffected(sourceID, targetID string) bool {
	if _, ok := g.lookupNode(sourceID); !ok {
		return false
	}
	if _, ok := g.lookupNode(targetID); !ok {
		return false
	}
	if sourceID == targetID {
		return true
	}

	adj := g.buildAdjacency()

	visited := make(map[string]bool)
	queue := []string{sourceID}
	visited[sourceID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighborID := range adj[current] {
			if neighborID == targetID {
				return true
			}
			if !visited[neighborID] {
				visited[neighborID] = true
				queue = append(queue, neighborID)
			}
		}
	}
	return false
}

func (g *Graph) buildAdjacency() map[string][]string {
	adj := make(map[string][]string)
	for _, e := range g.edges {
		adj[e.SourceID] = append(adj[e.SourceID], e.TargetID)
		adj[e.TargetID] = append(adj[e.TargetID], e.SourceID)
	}
	return adj
}
