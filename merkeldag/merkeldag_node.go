package merkeldag

import (
	"crypto/sha256"
	"errors"
)

// Node 表示 Merkle 树中的一个节点
type Node struct {
	left  *Node
	right *Node
	prev  *Node
	data  []byte
	hash  []byte
}

// NewNode 创建一个新的 Merkle 节点
func NewNode(data []byte) (*Node, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	hash := sha256.Sum256(data)
	return &Node{
		data:  data,
		hash:  hash[:],
		left:  nil,
		right: nil,
		prev:  nil,
	}, nil
}

// calculateHash 计算节点的哈希值
func (n *Node) calculateHash() []byte {
	if n == nil {
		return nil
	}

	// 如果是叶子节点，直接计算数据哈希
	if n.left == nil && n.right == nil {
		hash := sha256.Sum256(n.data)
		return hash[:]
	}

	// 如果是中间节点，计算子节点哈希的组合
	var combinedHash []byte
	if n.left != nil {
		combinedHash = append(combinedHash, n.left.hash...)
	}
	if n.right != nil {
		combinedHash = append(combinedHash, n.right.hash...)
	}

	hash := sha256.Sum256(combinedHash)
	return hash[:]
}

// updateHash 更新节点及其所有父节点的哈希值
func (n *Node) updateHash() {
	if n == nil {
		return
	}

	n.hash = n.calculateHash()
	if n.prev != nil {
		n.prev.updateHash()
	}
}

// InsertNode 在 Merkle 树中插入新节点
func InsertNode(root *Node, data []byte, index uint64) (*Node, error) {
	if root == nil {
		root, _ = NewNode([]byte("root"))
		// 初始化完整的空树结构
		initializeEmptyTree(root, TreeDepth)
	}

	// 插入叶子节点
	return insertAtIndex(root, data, index)
}

// initializeEmptyTree 初始化空的树结构
func initializeEmptyTree(node *Node, depth int) error {
	if depth == 0 {
		return nil
	}

	left, _ := NewNode([]byte("empty"))
	right, _ := NewNode([]byte("empty"))

	node.left = left
	node.right = right
	left.prev = node
	right.prev = node

	initializeEmptyTree(left, depth-1)
	initializeEmptyTree(right, depth-1)

	return nil
}

// insertAtIndex 在指定索引位置插入节点
func insertAtIndex(root *Node, data []byte, index uint64) (*Node, error) {
	if root == nil {
		return nil, errors.New("root cannot be nil")
	}

	// 创建新的叶子节点
	newNode, err := NewNode(data)
	if err != nil {
		return nil, err
	}

	// 找到目标叶子节点的路径
	current := root
	for i := TreeDepth - 1; i >= 0; i-- {
		if current == nil {
			return nil, errors.New("invalid tree structure")
		}

		// 使用index的第i位决定走左边还是右边
		if (index & (1 << uint(i))) == 0 {
			current = current.left
		} else {
			current = current.right
		}
	}

	// 替换叶子节点
	if current.prev.left == current {
		current.prev.left = newNode
	} else {
		current.prev.right = newNode
	}
	newNode.prev = current.prev

	// 更新哈希值
	newNode.updateHash()

	return GetRoot(newNode), nil
}

// GetRoot 获取树的根节点
func GetRoot(node *Node) *Node {
	if node == nil {
		return nil
	}

	current := node
	for current.prev != nil {
		current = current.prev
	}
	return current
}

// Verify 验证节点的哈希值是否正确
func (n *Node) Verify() bool {
	if n == nil {
		return true
	}

	expectedHash := n.calculateHash()
	if string(expectedHash) != string(n.hash) {
		return false
	}

	return n.left.Verify() && n.right.Verify()
}

// GetLeafNodes 获取所有叶子节点
func (n *Node) GetLeafNodes() []*Node {
	leaves := make([]*Node, 0)
	n.collectLeaves(&leaves)
	return leaves
}

// collectLeaves 收集叶子节点的辅助函数
func (n *Node) collectLeaves(leaves *[]*Node) {
	if n == nil {
		return
	}

	if n.left == nil && n.right == nil {
		*leaves = append(*leaves, n)
		return
	}

	n.left.collectLeaves(leaves)
	n.right.collectLeaves(leaves)
}

// Getter 方法
func (n *Node) GetHash() []byte {
	if n == nil {
		return nil
	}
	return n.hash
}

func (n *Node) GetData() []byte {
	if n == nil {
		return nil
	}
	return n.data
}

func (n *Node) GetLeft() *Node {
	if n == nil {
		return nil
	}
	return n.left
}

func (n *Node) GetRight() *Node {
	if n == nil {
		return nil
	}
	return n.right
}

func (n *Node) GetPrev() *Node {
	if n == nil {
		return nil
	}
	return n.prev
}

// Setter 方法
func (n *Node) SetLeft(left *Node) {
	if n == nil {
		return
	}
	n.left = left
	if left != nil {
		left.prev = n
	}
	n.updateHash()
}

func (n *Node) SetRight(right *Node) {
	if n == nil {
		return
	}
	n.right = right
	if right != nil {
		right.prev = n
	}
	n.updateHash()
}

func (n *Node) SetData(data []byte) error {
	if n == nil {
		return errors.New("node is nil")
	}
	if data == nil {
		return errors.New("data cannot be nil")
	}
	n.data = data
	n.updateHash()
	return nil
}
