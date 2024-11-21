package merkeldag

import (
	"errors"
	"fmt"
)

const TreeDepth = 8
const MaxLeafNodes = 1 << TreeDepth // 2^18 个叶子节点

type MerkelDAG struct {
	root         *Node
	currentIndex uint64 // 当前已使用的索引位置
}

// NewMerkelDAG 创建一个新的 Merkle 树
func NewMerkelDAG() (*MerkelDAG, error) {
	// 创建根节点
	root, err := NewNode([]byte("root"))
	if err != nil {
		return nil, fmt.Errorf("failed to create root node: %v", err)
	}

	// 初始化完整的空树结构
	if err := InitializeEmptyTree(root, TreeDepth); err != nil {
		return nil, fmt.Errorf("failed to initialize empty tree: %v", err)
	}

	return &MerkelDAG{
		root:         root,
		currentIndex: 0,
	}, nil
}

// initializeEmptyTree 从根节点开始初始化空的树结构
func InitializeEmptyTree(node *Node, depth int) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}

	if depth == 0 {
		// 叶子节点，计算哈希
		node.updateHash()
		return nil
	}

	// 创建子节点
	left, err := NewNode([]byte("empty"))
	if err != nil {
		return fmt.Errorf("failed to create left node: %v", err)
	}
	right, err := NewNode([]byte("empty"))
	if err != nil {
		return fmt.Errorf("failed to create right node: %v", err)
	}

	// 设置节点关系
	node.left = left
	node.right = right
	left.prev = node
	right.prev = node

	// 递归初始化子树
	if err := InitializeEmptyTree(left, depth-1); err != nil {
		return err
	}
	if err := InitializeEmptyTree(right, depth-1); err != nil {
		return err
	}

	// 计算当前节点的哈希
	node.updateHash()

	return nil
}

// Insert 插入新数据
func (m *MerkelDAG) Insert(data []byte) error {
	if m == nil {
		return errors.New("dag is nil")
	}

	if m.currentIndex >= MaxLeafNodes {
		return errors.New("merkle tree is full")
	}

	// 直接使用 insertAtIndex，因为树结构已经存在
	newRoot, err := insertAtIndex(m.root, data, m.currentIndex)
	if err != nil {
		return err
	}

	m.root = newRoot
	m.currentIndex++
	return nil
}

// GetRoot 获取 DAG 的根节点
func (m *MerkelDAG) GetRoot() *Node {
	if m == nil {
		return nil
	}
	return m.root
}

// GetCurrentIndex 获取当前索引
func (m *MerkelDAG) GetCurrentIndex() uint64 {
	if m == nil {
		return 0
	}
	return m.currentIndex
}

// SetRoot 设置根节点
func (m *MerkelDAG) SetRoot(root *Node) {
	if m == nil {
		return
	}
	m.root = root
}

// SetCurrentIndex 设置当前索引
func (m *MerkelDAG) SetCurrentIndex(index uint64) error {
	if m == nil {
		return errors.New("dag is nil")
	}
	if index >= MaxLeafNodes {
		return errors.New("index exceeds maximum leaf nodes")
	}
	m.currentIndex = index
	return nil
}

// IsEmpty 检查 DAG 是否为空
func (m *MerkelDAG) IsEmpty() bool {
	return m == nil || m.root == nil
}
