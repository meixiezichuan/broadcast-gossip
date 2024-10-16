package common

import (
	"container/list"
	"fmt"
)

// Graph represents an undirected graph using an adjacency list with string nodes
type Graph struct {
	adjList map[string][]string
	root    string
}

// NewGraph creates a new graph
func NewGraph() *Graph {
	return &Graph{
		adjList: make(map[string][]string),
		root:    "",
	}
}

// AddEdge adds an edge between two vertices
func (g *Graph) AddEdge(v1, v2 string) {
	if v1 == v2 {
		return
	}
	path1 := []string{v1, v2}
	path2 := []string{v2, v1}
	if g.PathExists(path1) || g.PathExists(path2) {
		return
	}
	g.adjList[v1] = append(g.adjList[v1], v2)
	g.adjList[v2] = append(g.adjList[v2], v1)
}

// RemoveEdge 删除两个顶点之间的边
func (g *Graph) RemoveEdge(v1, v2 string) {
	// 删除 v1 邻接表中与 v2 的边
	g.adjList[v1] = removeElement(g.adjList[v1], v2)
	// 删除 v2 邻接表中与 v1 的边
	g.adjList[v2] = removeElement(g.adjList[v2], v1)
}

// 辅助函数：从切片中删除指定元素
func removeElement(slice []string, element string) []string {
	for i, v := range slice {
		if v == element {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// FindMaxLeafTree finds the maximum leaf spanning tree starting from the given root
func (g *Graph) FindMaxLeafTree(root string) *Graph {
	tree := NewGraph()
	tree.root = root
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

func (g *Graph) FindNeighbor(node string) []string {
	return g.adjList[node]
}

// IsLeaf checks if a given node is a leaf in the tree
func (g *Graph) IsLeaf(node string) bool {
	// root node
	if g.root == node {
		return false
	}
	edges, exists := g.adjList[node]
	return exists && len(edges) == 1
}

func (g *Graph) PathExistsInTree(currentNode string, path []string) bool {
	if path[0] != currentNode {
		return false
	}
	// 如果路径长度为1，表示已经检查完所有节点，路径存在
	if len(path) == 1 {
		return true
	}

	// 获取当前节点的所有子节点
	children, exists := g.adjList[currentNode]
	if !exists {
		return false
	}

	// 检查路径的下一个节点是否是当前节点的子节点
	for _, child := range children {
		if child == path[1] {
			// 递归检查子节点的路径
			return g.PathExistsInTree(child, path[1:])
		}
	}

	return false
}

// PathExists checks if a given path exists in the tree
func (g *Graph) PathExists(path []string) bool {
	// If is a tree
	if g.root != "" {
		return g.PathExistsInTree(g.root, path)
	}

	// Not a tree
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
