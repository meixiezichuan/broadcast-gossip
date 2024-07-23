package gossip

import (
	"container/list"
	"fmt"
)

// Graph represents an undirected graph using an adjacency list with string nodes
type Graph struct {
	adjList map[string][]string
}

// NewGraph creates a new graph
func NewGraph() *Graph {
	return &Graph{
		adjList: make(map[string][]string),
	}
}

// AddEdge adds an edge between two vertices
func (g *Graph) AddEdge(v1, v2 string) {
	g.adjList[v1] = append(g.adjList[v1], v2)
	g.adjList[v2] = append(g.adjList[v2], v1)
}

// FindMaxLeafTree finds the maximum leaf spanning tree starting from the given root
func (g *Graph) FindMaxLeafTree(root string) *Graph {
	tree := NewGraph()
	visited := make(map[string]bool)
	queue := list.New()
	queue.PushBack(root)
	visited[root] = true

	for queue.Len() > 0 {
		node := queue.Remove(queue.Front()).(string)
		for _, neighbor := range g.adjList[node] {
			if !visited[neighbor] {
				tree.AddEdge(node, neighbor)
				queue.PushBack(neighbor)
				visited[neighbor] = true
			}
		}
	}
	return tree
}

// IsLeaf checks if a given node is a leaf in the tree
func (g *Graph) IsLeaf(node string) bool {
	edges, exists := g.adjList[node]
	return exists && len(edges) == 1
}

// PathExists checks if a given path exists in the tree
func (g *Graph) PathExists(path []string) bool {
	for i := 0; i < len(path)-1; i++ {
		found := false
		for _, neighbor := range g.adjList[path[i]] {
			if neighbor == path[i+1] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Display prints the adjacency list of the graph
func (g *Graph) Display() {
	for vertex, edges := range g.adjList {
		fmt.Printf("%s -> %v\n", vertex, edges)
	}
}
