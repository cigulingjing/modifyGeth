package merkeldag

import (
	"errors"

	"github.com/ethereum/go-ethereum/ethdb"
)

const TreeDepth = 8
const MaxLeafNodes = 1 << TreeDepth // 2^18 个叶子节点

type MerkelDAG struct {
	root         *Node
	currentIndex uint64 // 当前已使用的索引位置
}

// NewMerkelDAG 创建一个新的 Merkle 树
func NewMerkelDAG(db ethdb.Database) *MerkelDAG {
	return &MerkelDAG{
		root:         nil,
		currentIndex: 0,
	}
}

// Insert 插入新数据
func (m *MerkelDAG) Insert(data []byte) error {
	if m.currentIndex >= MaxLeafNodes {
		return errors.New("merkle tree is full")
	}

	newRoot, err := InsertNode(m.root, data, m.currentIndex)
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
