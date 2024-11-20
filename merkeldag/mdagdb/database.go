package mdagdb

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/merkeldag"
)

type MerkleDAGDB struct {
	db        ethdb.Database
	mu        sync.RWMutex
	nodeCache map[string]*merkeldag.Node
}

// NewMerkleDAGDB 创建新的数据库实例
func NewMerkleDAGDB(db ethdb.Database) *MerkleDAGDB {
	return &MerkleDAGDB{
		db:        db,
		nodeCache: make(map[string]*merkeldag.Node),
	}
}

// 序列化节点的结构
type serialNode struct {
	Data  []byte
	Hash  []byte
	Left  []byte // 存储左子节点的哈希
	Right []byte // 存储右子节点的哈希
}

// SaveDAG 保存整个 DAG
func (mdb *MerkleDAGDB) SaveDAG(dag *merkeldag.MerkelDAG) error {
	mdb.mu.Lock()
	defer mdb.mu.Unlock()

	if dag == nil || dag.GetRoot() == nil {
		return nil
	}

	// 保存当前索引
	if err := mdb.saveIndex(dag.GetCurrentIndex()); err != nil {
		return err
	}

	// 保存根节点哈希
	rootHash := dag.GetRoot().GetHash()
	if err := mdb.db.Put([]byte("root_hash"), rootHash); err != nil {
		return fmt.Errorf("failed to save root hash: %v", err)
	}

	// 保存所有节点
	return mdb.saveNode(dag.GetRoot())
}

// saveNode 递归保存节点
func (mdb *MerkleDAGDB) saveNode(node *merkeldag.Node) error {
	if node == nil {
		return nil
	}

	// 检查节点是否已存在
	hashStr := string(node.GetHash())
	if _, exists := mdb.nodeCache[hashStr]; exists {
		return nil
	}

	// 创建序列化结构
	sNode := &serialNode{
		Data: node.GetData(),
		Hash: node.GetHash(),
	}

	// 保存子节点引用
	if node.GetLeft() != nil {
		sNode.Left = node.GetLeft().GetHash()
		if err := mdb.saveNode(node.GetLeft()); err != nil {
			return err
		}
	}
	if node.GetRight() != nil {
		sNode.Right = node.GetRight().GetHash()
		if err := mdb.saveNode(node.GetRight()); err != nil {
			return err
		}
	}

	// 序列化并保存
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(sNode); err != nil {
		return err
	}

	if err := mdb.db.Put(node.GetHash(), buf.Bytes()); err != nil {
		return err
	}

	mdb.nodeCache[hashStr] = node
	return nil
}

// LoadDAG 从数据库加载 DAG
func (mdb *MerkleDAGDB) LoadDAG() (*merkeldag.MerkelDAG, error) {
	mdb.mu.Lock()
	defer mdb.mu.Unlock()

	// 清空缓存
	mdb.nodeCache = make(map[string]*merkeldag.Node)

	// 加载当前索引
	index, err := mdb.loadIndex()
	if err != nil {
		return nil, err
	}

	// 创建新的 DAG
	dag := merkeldag.NewMerkelDAG(mdb.db)

	// 如果索引为0，说明是空树
	if index == 0 {
		return dag, nil
	}

	// 加载根节点哈希
	rootHash, err := mdb.db.Get([]byte("root_hash"))
	if err != nil {
		return nil, fmt.Errorf("failed to load root hash: %v", err)
	}

	// 递归加载根节点及其所有子节点
	root, err := mdb.loadNode(rootHash)
	if err != nil {
		return nil, fmt.Errorf("failed to load root node: %v", err)
	}

	// 设置 DAG 的状态
	dag.SetRoot(root)
	if err := dag.SetCurrentIndex(index); err != nil {
		return nil, fmt.Errorf("failed to set current index: %v", err)
	}

	return dag, nil
}

// loadNode 递归加载节点
func (mdb *MerkleDAGDB) loadNode(hash []byte) (*merkeldag.Node, error) {
	if hash == nil {
		return nil, nil
	}

	// 检查缓存
	hashStr := string(hash)
	if node, exists := mdb.nodeCache[hashStr]; exists {
		return node, nil
	}

	// 从数据库加载
	data, err := mdb.db.Get(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to load node data: %v", err)
	}

	var sNode serialNode
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&sNode); err != nil {
		return nil, fmt.Errorf("failed to decode node: %v", err)
	}

	// 创建新节点
	node, err := merkeldag.NewNode(sNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %v", err)
	}

	// 添加到缓存（在加载子节点之前，防止循环引用）
	mdb.nodeCache[hashStr] = node

	// 递归加载左子节点
	if sNode.Left != nil {
		left, err := mdb.loadNode(sNode.Left)
		if err != nil {
			return nil, fmt.Errorf("failed to load left child: %v", err)
		}
		node.SetLeft(left)
	}

	// 递归加载右子节点
	if sNode.Right != nil {
		right, err := mdb.loadNode(sNode.Right)
		if err != nil {
			return nil, fmt.Errorf("failed to load right child: %v", err)
		}
		node.SetRight(right)
	}

	return node, nil
}

// 保存当前索引
func (mdb *MerkleDAGDB) saveIndex(index uint64) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(index); err != nil {
		return err
	}
	return mdb.db.Put([]byte("current_index"), buf.Bytes())
}

// 加载当前索引
func (mdb *MerkleDAGDB) loadIndex() (uint64, error) {
	data, err := mdb.db.Get([]byte("current_index"))
	if err != nil {
		return 0, nil // 如果不存在，返回0
	}

	var index uint64
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&index); err != nil {
		return 0, err
	}
	return index, nil
}

// ClearCache 清除缓存
func (mdb *MerkleDAGDB) ClearCache() {
	mdb.mu.Lock()
	defer mdb.mu.Unlock()
	mdb.nodeCache = make(map[string]*merkeldag.Node)
}
