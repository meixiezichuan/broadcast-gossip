package common

import (
	"container/list"
	"fmt"
	"sort"
)

// Graph represents an undirected graph using an adjacency list with string nodes
type Graph struct {
	adjList map[string][]string
	root    string
}

type DPState struct {
	notIncluded int // 不包括当前节点的最多叶子数
	included    int // 包括当前节点的最多叶子数
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

// GetSortedNodes returns nodes sorted by degree and name
func (g *Graph) GetSortedNodes(nodes []string) []string {
	nodeDegrees := make(map[string]int)
	for _, node := range nodes {
		neighbors := g.adjList[node]
		nodeDegrees[node] = len(neighbors)
	}
	// Convert map to slice and sort
	ns := make([]string, 0, len(nodeDegrees))

	for key, _ := range nodeDegrees {
		ns = append(ns, key)
	}
	sort.Slice(ns, func(i, j int) bool { return nodeDegrees[ns[i]] > nodeDegrees[ns[j]] })

	//fmt.Println("sotr nodes : ", nodes, "------------")
	//for _, n := range ns {
	//	fmt.Println("Node: ", n, "Degree: ", nodeDegrees[n])
	//}
	return ns
}

// MinDominatingSetFromRoot 查找以指定根节点为起始的最小支配集
func (g *Graph) MinDominatingSetFromRoot(root string) []string {
	dominatingSet := []string{}
	covered := make(map[string]bool)

	// 首先将根节点添加到支配集中
	dominatingSet = append(dominatingSet, root)
	covered[root] = true

	// 标记根节点的邻居为已覆盖
	for _, neighbor := range g.adjList[root] {
		covered[neighbor] = true
	}

	var nodes []string
	for n, _ := range g.adjList {
		nodes = append(nodes, n)
	}

	sortedNodes := g.GetSortedNodes(nodes)
	// 循环直到所有节点都被覆盖
	for len(covered) < len(g.adjList) {
		var maxCoverNode string
		maxCoverCount := -1

		// 找到能覆盖最多未覆盖节点的顶点
		for _, node := range sortedNodes {
			if covered[node] {
				continue
			}

			// 计算未覆盖邻居的数量
			coverCount := 0
			for _, neighbor := range g.adjList[node] {
				if !covered[neighbor] {
					coverCount++
				}
			}

			// 更新最大覆盖的节点
			if coverCount > maxCoverCount {
				maxCoverCount = coverCount
				maxCoverNode = node
			}
		}

		// 将选择的节点添加到支配集，并标记其邻居为已覆盖
		if maxCoverNode != "" {
			dominatingSet = append(dominatingSet, maxCoverNode)
			covered[maxCoverNode] = true
			for _, neighbor := range g.adjList[maxCoverNode] {
				covered[neighbor] = true
			}
		}
	}

	return dominatingSet
}

func (g *Graph) MaxLeafSpanningTree(root string) (*Graph, int, []string) {

	dp := make(map[string]DPState) // DP表
	//parent := make(map[string]string) // 记录父节点
	//
	//var dfs func(node, p string)
	//dfs = func(node, p string) {
	//	dp[node] = DPState{0, 1} // 初始化包含当前节点
	//	parent[node] = p         // 记录父节点
	//	for _, neighbor := range g.adjList[node] {
	//		if neighbor == p {
	//			continue // 避免回到父节点
	//		}
	//		dfs(neighbor, node)
	//		notIcud := dp[node].notIncluded + dp[neighbor].included
	//		icud := dp[node].included + max(dp[neighbor].notIncluded, dp[neighbor].included)
	//
	//		dp[node] = DPState{notIncluded: notIcud, included: icud} // 不包括当前节点，包含子节点
	//	}
	//}
	//
	//dfs(root, "")
	g.DFS(root, dp)

	// 构建生成树和叶子节点
	tree := make(map[string][]string)
	var maxLeaves int
	if dp[root].notIncluded > dp[root].included {
		maxLeaves = dp[root].notIncluded
	} else {
		maxLeaves = dp[root].included
	}

	// 记录叶子节点
	var leafNodes []string
	visited := make(map[string]bool)

	var buildTree func(node, parent string, include bool)
	buildTree = func(node, parent string, include bool) {
		if include {
			if visited[node] {
				return
			}
			if parent != "" {
				tree[parent] = append(tree[parent], node) // 添加到树结构
			}
			isLeaf := true
			visited[node] = true
			for _, neighbor := range g.adjList[node] {
				if neighbor == parent {
					continue
				}
				isLeaf = false
				// 根据DP状态选择是否包括子节点
				if dp[neighbor].included >= dp[neighbor].notIncluded {
					buildTree(neighbor, node, true) // 包括子节点
				} else {
					buildTree(neighbor, node, false) // 不包括子节点
				}
			}
			if isLeaf {
				leafNodes = append(leafNodes, node) // 记录叶子节点
			}
		}
	}
	buildTree(root, "", dp[root].included >= dp[root].notIncluded)

	mlstree := &Graph{
		adjList: tree,
		root:    root,
	}

	return mlstree, maxLeaves, leafNodes
}

func (g *Graph) DFS(root string, dp map[string]DPState) {
	var stack []struct {
		node   string
		parent string
	}
	stack = append(stack, struct {
		node   string
		parent string
	}{node: root, parent: ""})

	// DFS 遍历
	visited := make(map[string]bool)
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1] // 弹出栈顶元素
		node := top.node
		parent := top.parent

		if visited[node] {
			continue // 如果节点已访问，则跳过
		}
		visited[node] = true // 标记为已访问

		fmt.Println("node: ", node)
		dp[node] = DPState{0, 1} // 初始化包含当前节点
		for _, neighbor := range g.adjList[node] {
			if neighbor == parent {
				continue // 避免回到父节点
			}
			stack = append(stack, struct {
				node   string
				parent string
			}{node: neighbor, parent: node}) // 将邻居添加到栈中
		}
	}

	// 从后往前处理每个节点
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1] // 弹出栈顶元素
		node := top.node
		parent := top.parent

		for _, neighbor := range g.adjList[node] {
			if neighbor == parent {
				continue // 避免回到父节点
			}

			stack = append(stack, struct {
				node   string
				parent string
			}{node: neighbor, parent: node}) // 将邻居添加到栈中
			notIcud := dp[node].notIncluded + dp[neighbor].included
			icud := dp[node].included + max(dp[neighbor].notIncluded, dp[neighbor].included)

			dp[node] = DPState{notIncluded: notIcud, included: icud}
		}
	}
}

func (g *Graph) BuildMDSTree(root string) *Graph {
	mds := g.MinDominatingSetFromRoot(root)
	fmt.Println("mds set: ", mds)

	covered := make(map[string]bool)

	// 将最小支配集中的所有节点标记为已覆盖
	for _, node := range mds {
		covered[node] = true
	}

	tree := NewGraph()
	tree.root = root

	visited := make(map[string]bool)

	// 使用深度优先搜索（DFS）构建生成树
	var dfs func(node, parent string)
	dfs = func(node, parent string) {
		visited[node] = true
		if parent != "" {
			tree.AddEdge(parent, node)
			//tree[parent] = append(tree[parent], node)
		}
		for _, neighbor := range g.adjList[node] {
			if !visited[neighbor] {
				tree.AddEdge(neighbor, node)
				dfs(neighbor, node)
			}
		}
	}

	// 从根节点开始构建生成树
	dfs(root, "")

	return tree
}

func (g *Graph) BuildSpanningTree(mds []string) map[string][]string {
	tree := make(map[string][]string)
	visited := make(map[string]bool)

	// 深度优先搜索（DFS）构建生成树
	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		for _, neighbor := range g.adjList[node] {
			if !visited[neighbor] {
				tree[node] = append(tree[node], neighbor)
				tree[neighbor] = append(tree[neighbor], node)
				dfs(neighbor)
			}
		}
	}

	// 遍历最小支配集，构建生成树
	for _, node := range mds {
		if !visited[node] {
			dfs(node)
		}
	}

	return tree
}

func (g *Graph) BST(root string) *Graph {
	mds := g.MinDominatingSetFromRoot(root)
	adj := g.BuildSpanningTree(mds)
	tree := NewGraph()
	tree.root = root
	tree.adjList = adj
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
	//if g.root != "" {
	//	return g.PathExistsInTree(g.root, path)
	//}

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

// DPState 定义状态
type DPMState struct {
	maxLeaves int
	leafNodes []string
}

// MaxLeafSpanningTree 计算最大叶子生成树
func (g *Graph) MLST2DFS(root string) (int, map[string][]string, []string) {
	dp := make(map[string]DPMState)   // DP表
	tree := make(map[string][]string) // 记录生成树结构
	visited := make(map[string]bool)

	// DFS 遍历并更新状态
	var dfs func(node, parent string) DPMState
	dfs = func(node, parent string) DPMState {
		visited[node] = true
		state := DPMState{0, nil} // 初始化状态
		isLeaf := true            // 假设当前节点是叶子节点

		for _, neighbor := range g.adjList[node] {
			if neighbor == parent || visited[neighbor] {
				continue // 跳过父节点
			}
			childState := dfs(neighbor, node)         // 递归访问邻居
			state.maxLeaves += childState.maxLeaves   // 更新叶子数量
			tree[node] = append(tree[node], neighbor) // 记录树结构

			isLeaf = false // 发现子节点，当前节点不是叶子节点
		}

		if isLeaf {
			state.maxLeaves = 1                             // 叶子节点自身计数
			state.leafNodes = append(state.leafNodes, node) // 记录叶子节点
		} else {
			// 不是叶子节点，按需更新
			state.leafNodes = append(state.leafNodes, state.leafNodes...) // 将所有子节点的叶子节点合并
		}

		dp[node] = state // 保存当前节点状态
		return state
	}

	// 开始 DFS
	dfs(root, "")

	// 返回结果
	return dp[root].maxLeaves, tree, dp[root].leafNodes
}

// MaxLeafSpanningTree 计算最大叶子生成树
func (g *Graph) MLSTBFS(root string) (int, map[string][]string, []string) {
	dp := make(map[string]DPMState)   // DP表
	tree := make(map[string][]string) // 记录生成树结构
	visited := make(map[string]bool)  // 记录已访问节点
	queue := []string{root}           // BFS 队列
	visited[root] = true              // 标记根节点为已访问

	for len(queue) > 0 {
		// 当前层的节点数量
		levelSize := len(queue)
		levelStates := make(map[string]DPMState) // 当前层的状态

		for i := 0; i < levelSize; i++ {
			node := queue[0]
			queue = queue[1:] // 出队

			state := DPMState{0, nil} // 初始化状态
			isLeaf := true            // 假设当前节点是叶子节点

			// 获取并排序邻接节点，确保顺序一致
			neighbors := g.adjList[node]
			sort.Strings(neighbors) // 排序邻接节点

			for _, neighbor := range neighbors {
				if visited[neighbor] {
					continue // 跳过已访问的节点
				}
				queue = append(queue, neighbor) // 入队
				visited[neighbor] = true        // 标记为已访问
				isLeaf = false                  // 当前节点不是叶子节点

				// 初始化邻接节点的状态
				if _, exists := levelStates[neighbor]; !exists {
					levelStates[neighbor] = DPMState{0, nil}
				}
				childState := levelStates[neighbor]
				state.maxLeaves += childState.maxLeaves   // 更新叶子数量
				tree[node] = append(tree[node], neighbor) // 记录树结构
			}

			if isLeaf {
				state.maxLeaves = 1                             // 叶子节点自身计数
				state.leafNodes = append(state.leafNodes, node) // 记录叶子节点
			} else {
				// 不是叶子节点，合并子节点的叶子节点
				for _, child := range state.leafNodes {
					state.leafNodes = append(state.leafNodes, child)
				}
			}

			levelStates[node] = state // 保存当前节点状态
		}

		// 更新 dp 表
		for _, state := range levelStates {
			if len(state.leafNodes) > 0 {
				dp[state.leafNodes[0]] = state // 保存当前层的状态
			}
		}
	}

	// 查找最大叶子数量
	maxLeaves := 0
	for _, state := range dp {
		if state.maxLeaves > maxLeaves {
			maxLeaves = state.maxLeaves
		}
	}

	return maxLeaves, tree, dp[root].leafNodes
}

// Display prints the adjacency list of the graph
func (g *Graph) Display() {
	for vertex, edges := range g.adjList {
		fmt.Printf("%s -> %v\n", vertex, edges)
	}
}

func (g *Graph) Sotred() {
	fmt.Println("Before sorted: ---")
	g.Display()
	for vertex, edges := range g.adjList {
		g.adjList[vertex] = g.GetSortedNodes(edges)
	}
	fmt.Println("After sorted: ---")
	g.Display()
}
